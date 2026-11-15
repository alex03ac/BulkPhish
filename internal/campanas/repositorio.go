package campanas

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"bulkphish/internal/eventos"
)

// repositorio concentra todo el acceso a la base de datos (esquema "campanas").
type repositorio struct {
	bd *pgxpool.Pool
}

var (
	errCorreoDuplicado  = errors.New("correo duplicado")
	errSinDestinatarios = errors.New("sin destinatarios")
)

// ---------- Empleados ----------

type Empleado struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Correo string `json:"correo"`
	Area   string `json:"area"`
}

type AreaConteo struct {
	Area      string `json:"area"`
	Empleados int    `json:"empleados"`
}

func (r *repositorio) listarEmpleados(ctx context.Context) ([]Empleado, error) {
	filas, err := r.bd.Query(ctx, `SELECT id::text, nombre, correo, area FROM empleados ORDER BY area, nombre`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	lista := []Empleado{}
	for filas.Next() {
		var e Empleado
		if err := filas.Scan(&e.ID, &e.Nombre, &e.Correo, &e.Area); err != nil {
			return nil, err
		}
		lista = append(lista, e)
	}
	return lista, filas.Err()
}

func (r *repositorio) crearEmpleado(ctx context.Context, e Empleado) (Empleado, error) {
	err := r.bd.QueryRow(ctx,
		`INSERT INTO empleados (nombre, correo, area) VALUES ($1, lower($2), $3) RETURNING id::text, correo`,
		e.Nombre, e.Correo, e.Area).Scan(&e.ID, &e.Correo)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // violación de unicidad
		return Empleado{}, errCorreoDuplicado
	}
	return e, err
}

func (r *repositorio) listarAreas(ctx context.Context) ([]AreaConteo, error) {
	filas, err := r.bd.Query(ctx, `SELECT area, count(*) FROM empleados GROUP BY area ORDER BY area`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	lista := []AreaConteo{}
	for filas.Next() {
		var a AreaConteo
		if err := filas.Scan(&a.Area, &a.Empleados); err != nil {
			return nil, err
		}
		lista = append(lista, a)
	}
	return lista, filas.Err()
}

// ---------- Campañas ----------

type Campana struct {
	ID                 string    `json:"id"`
	Nombre             string    `json:"nombre"`
	PlantillaID        string    `json:"plantilla_id"`
	CreadaPor          string    `json:"creada_por"`
	CreadaEn           time.Time `json:"creada_en"`
	TotalDestinatarios int       `json:"total_destinatarios"`
	Areas              []string  `json:"areas"`
}

func generarToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b) // 32 caracteres hexadecimales imposibles de adivinar
}

// crearCampana guarda la campaña, sus destinatarios y el evento "campana.creada"
// en UNA sola transacción. El evento se publica después, desde la bandeja de salida.
func (r *repositorio) crearCampana(ctx context.Context, usuario, nombre, plantillaID string, areas []string) (Campana, error) {
	tx, err := r.bd.Begin(ctx)
	if err != nil {
		return Campana{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Empleados objetivo (si no se eligen áreas, todos).
	var filas pgx.Rows
	if len(areas) == 0 {
		filas, err = tx.Query(ctx, `SELECT id::text, nombre, correo, area FROM empleados ORDER BY area, nombre`)
	} else {
		filas, err = tx.Query(ctx, `SELECT id::text, nombre, correo, area FROM empleados WHERE area = ANY($1) ORDER BY area, nombre`, areas)
	}
	if err != nil {
		return Campana{}, err
	}
	objetivo := []Empleado{}
	for filas.Next() {
		var e Empleado
		if err := filas.Scan(&e.ID, &e.Nombre, &e.Correo, &e.Area); err != nil {
			filas.Close()
			return Campana{}, err
		}
		objetivo = append(objetivo, e)
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return Campana{}, err
	}
	if len(objetivo) == 0 {
		return Campana{}, errSinDestinatarios
	}

	// 2. La campaña.
	c := Campana{Nombre: nombre, PlantillaID: plantillaID, CreadaPor: usuario, TotalDestinatarios: len(objetivo)}
	err = tx.QueryRow(ctx,
		`INSERT INTO campanas (nombre, plantilla_id, creada_por) VALUES ($1, $2, $3) RETURNING id::text, creada_en`,
		nombre, plantillaID, usuario).Scan(&c.ID, &c.CreadaEn)
	if err != nil {
		return Campana{}, err
	}

	// 3. Un destinatario (con su token único) por empleado.
	destinatarios := make([]eventos.DestinatarioCampana, 0, len(objetivo))
	areasVistas := map[string]bool{}
	for _, e := range objetivo {
		d := eventos.DestinatarioCampana{Token: generarToken(), Nombre: e.Nombre, Correo: e.Correo, Area: e.Area}
		err := tx.QueryRow(ctx,
			`INSERT INTO destinatarios (campana_id, empleado_id, token) VALUES ($1, $2, $3) RETURNING id::text`,
			c.ID, e.ID, d.Token).Scan(&d.DestinatarioID)
		if err != nil {
			return Campana{}, err
		}
		destinatarios = append(destinatarios, d)
		if !areasVistas[e.Area] {
			areasVistas[e.Area] = true
			c.Areas = append(c.Areas, e.Area)
		}
	}

	// 4. El evento, en la bandeja de salida (misma transacción).
	sobre, err := eventos.NuevoSobreCampana(c.ID, nombre, plantillaID, c.CreadaEn, destinatarios)
	if err != nil {
		return Campana{}, err
	}
	if err := encolarEvento(ctx, tx, eventos.CampanaCreada, sobre); err != nil {
		return Campana{}, err
	}
	return c, tx.Commit(ctx)
}

func (r *repositorio) listarCampanas(ctx context.Context) ([]Campana, error) {
	filas, err := r.bd.Query(ctx, `
		SELECT c.id::text, c.nombre, c.plantilla_id, c.creada_por, c.creada_en,
		       count(d.id), coalesce(array_agg(DISTINCT e.area) FILTER (WHERE e.area IS NOT NULL), '{}')
		FROM campanas c
		LEFT JOIN destinatarios d ON d.campana_id = c.id
		LEFT JOIN empleados e ON e.id = d.empleado_id
		GROUP BY c.id ORDER BY c.creada_en DESC`)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	lista := []Campana{}
	for filas.Next() {
		var c Campana
		if err := filas.Scan(&c.ID, &c.Nombre, &c.PlantillaID, &c.CreadaPor, &c.CreadaEn, &c.TotalDestinatarios, &c.Areas); err != nil {
			return nil, err
		}
		lista = append(lista, c)
	}
	return lista, filas.Err()
}

func (r *repositorio) obtenerCampana(ctx context.Context, id string) (Campana, error) {
	var c Campana
	err := r.bd.QueryRow(ctx, `
		SELECT c.id::text, c.nombre, c.plantilla_id, c.creada_por, c.creada_en,
		       count(d.id), coalesce(array_agg(DISTINCT e.area) FILTER (WHERE e.area IS NOT NULL), '{}')
		FROM campanas c
		LEFT JOIN destinatarios d ON d.campana_id = c.id
		LEFT JOIN empleados e ON e.id = d.empleado_id
		WHERE c.id = $1 GROUP BY c.id`, id).
		Scan(&c.ID, &c.Nombre, &c.PlantillaID, &c.CreadaPor, &c.CreadaEn, &c.TotalDestinatarios, &c.Areas)
	return c, err
}

// ---------- Interacciones públicas (clic y reporte) ----------

// registrarInteraccion marca el primer clic (o el primer reporte) de un destinatario y
// guarda el evento correspondiente en la bandeja de salida. Si ya estaba registrado,
// no hace nada: así repetir el clic no genera eventos duplicados.
// Devuelve false si el token no existe.
func (r *repositorio) registrarInteraccion(ctx context.Context, token, tipo string) (bool, error) {
	tx, err := r.bd.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var consulta, clave string
	switch tipo {
	case "clic":
		consulta = `UPDATE destinatarios SET clic_en = now() WHERE token = $1 AND clic_en IS NULL RETURNING id::text, campana_id::text`
		clave = eventos.EnlaceClicado
	case "reporte":
		consulta = `UPDATE destinatarios SET reportado_en = now() WHERE token = $1 AND reportado_en IS NULL RETURNING id::text, campana_id::text`
		clave = eventos.CorreoReportado
	default:
		return false, errors.New("tipo de interacción desconocido")
	}

	var destinatarioID, campanaID string
	err = tx.QueryRow(ctx, consulta, token).Scan(&destinatarioID, &campanaID)
	if errors.Is(err, pgx.ErrNoRows) {
		// O el token no existe o la interacción ya estaba registrada.
		var existe bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM destinatarios WHERE token = $1)`, token).Scan(&existe); err != nil {
			return false, err
		}
		return existe, nil
	}
	if err != nil {
		return false, err
	}

	sobre, err := eventos.NuevoSobre(clave, eventos.DatosInteraccion{
		CampanaID: campanaID, DestinatarioID: destinatarioID, OcurridoEn: time.Now().UTC(),
	})
	if err != nil {
		return false, err
	}
	if err := encolarEvento(ctx, tx, clave, sobre); err != nil {
		return false, err
	}
	return true, tx.Commit(ctx)
}

// InfoCapacitacion es lo que ve la persona en la página de capacitación.
type InfoCapacitacion struct {
	Nombre      string `json:"nombre"`
	PlantillaID string `json:"plantilla_id"`
	Clic        bool   `json:"clic"`
	Reportado   bool   `json:"reportado"`
}

func (r *repositorio) infoCapacitacion(ctx context.Context, token string) (InfoCapacitacion, error) {
	var i InfoCapacitacion
	err := r.bd.QueryRow(ctx, `
		SELECT e.nombre, c.plantilla_id, d.clic_en IS NOT NULL, d.reportado_en IS NOT NULL
		FROM destinatarios d
		JOIN empleados e ON e.id = d.empleado_id
		JOIN campanas c ON c.id = d.campana_id
		WHERE d.token = $1`, token).Scan(&i.Nombre, &i.PlantillaID, &i.Clic, &i.Reportado)
	return i, err
}

// ---------- Bandeja de salida (outbox) ----------

type ejecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// encolarEvento guarda el evento en la tabla bandeja_salida dentro de la transacción actual.
func encolarEvento(ctx context.Context, tx ejecutor, clave string, sobre eventos.Sobre) error {
	crudo, err := json.Marshal(sobre)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO bandeja_salida (id, clave_ruta, sobre) VALUES ($1, $2, $3::jsonb)`,
		sobre.ID, clave, string(crudo))
	return err
}
