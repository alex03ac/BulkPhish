package resultados

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"bulkphish/internal/eventos"
)

// repositorio guarda y consulta el "modelo de lectura" (esquema "resultados").
// Es una copia propia de lo que este servicio necesita, construida a partir de eventos.
type repositorio struct {
	bd *pgxpool.Pool
}

var errAunNoExiste = errors.New("el destinatario aún no existe; se reintentará")

// ---------- Escritura (desde eventos) ----------

func (r *repositorio) registrarCampana(ctx context.Context, d eventos.DatosCampanaCreada) error {
	tx, err := r.bd.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx,
		`INSERT INTO campanas (campana_id, nombre, plantilla_id, creada_en) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (campana_id) DO NOTHING`,
		d.CampanaID, d.Nombre, d.PlantillaID, d.CreadaEn); err != nil {
		return err
	}
	for _, x := range d.Destinatarios {
		if _, err := tx.Exec(ctx,
			`INSERT INTO destinatarios (destinatario_id, campana_id, nombre, correo, area) VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (destinatario_id) DO NOTHING`,
			x.DestinatarioID, d.CampanaID, x.Nombre, x.Correo, x.Area); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// marcar fija la fecha de un hito (enviado, clic o reporte) solo la primera vez,
// de modo que recibir el mismo evento dos veces no cambia nada (idempotencia).
func (r *repositorio) marcar(ctx context.Context, columna, destinatarioID string, cuando time.Time) error {
	var consulta string
	switch columna {
	case "enviado_en":
		consulta = `UPDATE destinatarios SET enviado_en = coalesce(enviado_en, $2) WHERE destinatario_id = $1`
	case "clic_en":
		consulta = `UPDATE destinatarios SET clic_en = coalesce(clic_en, $2) WHERE destinatario_id = $1`
	case "reportado_en":
		consulta = `UPDATE destinatarios SET reportado_en = coalesce(reportado_en, $2) WHERE destinatario_id = $1`
	default:
		return errors.New("columna desconocida")
	}
	etiqueta, err := r.bd.Exec(ctx, consulta, destinatarioID, cuando)
	if err != nil {
		return err
	}
	if etiqueta.RowsAffected() == 0 {
		return errAunNoExiste // el evento llegó antes que "campana.creada": reintentar
	}
	return nil
}

// ---------- Lectura (para el panel) ----------

type Estadisticas struct {
	Enviados    int     `json:"enviados"`
	Clics       int     `json:"clics"`
	Reportados  int     `json:"reportados"`
	Total       int     `json:"total"`
	TasaClic    float64 `json:"tasa_clic"`    // clics / enviados (0 a 1)
	TasaReporte float64 `json:"tasa_reporte"` // reportados / enviados (0 a 1)
}

func calcular(total, enviados, clics, reportados int) Estadisticas {
	e := Estadisticas{Total: total, Enviados: enviados, Clics: clics, Reportados: reportados}
	if enviados > 0 {
		e.TasaClic = float64(clics) / float64(enviados)
		e.TasaReporte = float64(reportados) / float64(enviados)
	}
	return e
}

type ResumenCampana struct {
	ID          string    `json:"id"`
	Nombre      string    `json:"nombre"`
	PlantillaID string    `json:"plantilla_id"`
	CreadaEn    time.Time `json:"creada_en"`
	Estadisticas
}

type Grupo struct {
	Clave string `json:"clave"`
	Estadisticas
}

type Destinatario struct {
	Nombre      string     `json:"nombre"`
	Correo      string     `json:"correo"`
	Area        string     `json:"area"`
	Estado      string     `json:"estado"` // pendiente | enviado | clic | reportado
	EnviadoEn   *time.Time `json:"enviado_en"`
	ClicEn      *time.Time `json:"clic_en"`
	ReportadoEn *time.Time `json:"reportado_en"`
}

func (r *repositorio) campanas(ctx context.Context, limite int) ([]ResumenCampana, error) {
	filas, err := r.bd.Query(ctx, `
		SELECT c.campana_id::text, c.nombre, c.plantilla_id, c.creada_en,
		       count(d.destinatario_id), count(d.enviado_en), count(d.clic_en), count(d.reportado_en)
		FROM campanas c LEFT JOIN destinatarios d ON d.campana_id = c.campana_id
		GROUP BY c.campana_id ORDER BY c.creada_en DESC LIMIT $1`, limite)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	lista := []ResumenCampana{}
	for filas.Next() {
		var c ResumenCampana
		var total, env, cl, rep int
		if err := filas.Scan(&c.ID, &c.Nombre, &c.PlantillaID, &c.CreadaEn, &total, &env, &cl, &rep); err != nil {
			return nil, err
		}
		c.Estadisticas = calcular(total, env, cl, rep)
		lista = append(lista, c)
	}
	return lista, filas.Err()
}

func (r *repositorio) campana(ctx context.Context, id string) (*ResumenCampana, error) {
	lista, err := r.campanasPorID(ctx, id)
	if err != nil || len(lista) == 0 {
		return nil, err
	}
	return &lista[0], nil
}

func (r *repositorio) campanasPorID(ctx context.Context, id string) ([]ResumenCampana, error) {
	filas, err := r.bd.Query(ctx, `
		SELECT c.campana_id::text, c.nombre, c.plantilla_id, c.creada_en,
		       count(d.destinatario_id), count(d.enviado_en), count(d.clic_en), count(d.reportado_en)
		FROM campanas c LEFT JOIN destinatarios d ON d.campana_id = c.campana_id
		WHERE c.campana_id = $1 GROUP BY c.campana_id`, id)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	lista := []ResumenCampana{}
	for filas.Next() {
		var c ResumenCampana
		var total, env, cl, rep int
		if err := filas.Scan(&c.ID, &c.Nombre, &c.PlantillaID, &c.CreadaEn, &total, &env, &cl, &rep); err != nil {
			return nil, err
		}
		c.Estadisticas = calcular(total, env, cl, rep)
		lista = append(lista, c)
	}
	return lista, filas.Err()
}

// porArea agrupa por área; si campanaID no está vacío, limita a esa campaña.
func (r *repositorio) porArea(ctx context.Context, campanaID string) ([]Grupo, error) {
	filas, err := r.bd.Query(ctx, `
		SELECT area, count(*), count(enviado_en), count(clic_en), count(reportado_en)
		FROM destinatarios WHERE ($1 = '' OR campana_id::text = $1)
		GROUP BY area ORDER BY area`, campanaID)
	if err != nil {
		return nil, err
	}
	return leerGrupos(filas)
}

func (r *repositorio) porPlantilla(ctx context.Context) ([]Grupo, error) {
	filas, err := r.bd.Query(ctx, `
		SELECT c.plantilla_id, count(*), count(d.enviado_en), count(d.clic_en), count(d.reportado_en)
		FROM destinatarios d JOIN campanas c ON c.campana_id = d.campana_id
		GROUP BY c.plantilla_id ORDER BY c.plantilla_id`)
	if err != nil {
		return nil, err
	}
	return leerGrupos(filas)
}

type lectorFilas interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

func leerGrupos(filas lectorFilas) ([]Grupo, error) {
	defer filas.Close()
	lista := []Grupo{}
	for filas.Next() {
		var g Grupo
		var total, env, cl, rep int
		if err := filas.Scan(&g.Clave, &total, &env, &cl, &rep); err != nil {
			return nil, err
		}
		g.Estadisticas = calcular(total, env, cl, rep)
		lista = append(lista, g)
	}
	return lista, filas.Err()
}

func (r *repositorio) destinatarios(ctx context.Context, campanaID string) ([]Destinatario, error) {
	filas, err := r.bd.Query(ctx, `
		SELECT nombre, correo, area, enviado_en, clic_en, reportado_en
		FROM destinatarios WHERE campana_id = $1 ORDER BY area, nombre`, campanaID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	lista := []Destinatario{}
	for filas.Next() {
		var d Destinatario
		if err := filas.Scan(&d.Nombre, &d.Correo, &d.Area, &d.EnviadoEn, &d.ClicEn, &d.ReportadoEn); err != nil {
			return nil, err
		}
		switch {
		case d.ClicEn != nil:
			d.Estado = "clic"
		case d.ReportadoEn != nil:
			d.Estado = "reportado"
		case d.EnviadoEn != nil:
			d.Estado = "enviado"
		default:
			d.Estado = "pendiente"
		}
		lista = append(lista, d)
	}
	return lista, filas.Err()
}
