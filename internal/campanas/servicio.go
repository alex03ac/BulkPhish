// Paquete campanas: núcleo de negocio. Gestiona empleados y campañas, registra los clics
// y reportes, y publica los eventos en RabbitMQ (vía bandeja de salida).
package campanas

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5"

	"bulkphish/internal/basedatos"
	"bulkphish/internal/config"
	"bulkphish/internal/mensajeria"
	"bulkphish/internal/plantillas"
	"bulkphish/internal/web"
)

type servidor struct {
	repo *repositorio
}

// Ejecutar arranca el servicio y bloquea hasta recibir una señal de apagado.
func Ejecutar() error {
	config.IniciarRegistro("campanas")
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	bd, err := basedatos.Conectar(ctx, config.Obligatorio("BD_URL"))
	if err != nil {
		return err
	}
	defer bd.Close()

	mq, err := mensajeria.Conectar(ctx, config.Obligatorio("AMQP_URL"))
	if err != nil {
		return err
	}
	defer mq.Cerrar()

	repo := &repositorio{bd: bd}
	go (&relevo{repo: repo, mq: mq}).ejecutar(ctx)

	s := &servidor{repo: repo}
	return web.Servir(ctx, config.Texto("PUERTO", "8082"), s.rutas())
}

func (s *servidor) rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /salud", web.Salud)

	// Rutas protegidas: NGINX exige un JWT válido antes de llegar aquí.
	mux.HandleFunc("GET /api/plantillas", s.listarPlantillas)
	mux.HandleFunc("GET /api/areas", s.listarAreas)
	mux.HandleFunc("GET /api/empleados", s.listarEmpleados)
	mux.HandleFunc("POST /api/empleados", s.crearEmpleado)
	mux.HandleFunc("GET /api/campanas", s.listarCampanas)
	mux.HandleFunc("POST /api/campanas", s.crearCampana)
	mux.HandleFunc("GET /api/campanas/{id}", s.obtenerCampana)

	// Rutas públicas: las abre el empleado desde su correo; no llevan JWT.
	mux.HandleFunc("GET /t/{token}", s.clic)
	mux.HandleFunc("GET /r/{token}", s.reporte)
	mux.HandleFunc("GET /api/publico/capacitacion/{token}", s.capacitacion)
	return mux
}

func errorInterno(w http.ResponseWriter, contexto string, err error) {
	slog.Error(contexto, "error", err)
	web.Error(w, http.StatusInternalServerError, "Ocurrió un error inesperado. Inténtalo de nuevo.")
}

// ---------- Plantillas ----------

func (s *servidor) listarPlantillas(w http.ResponseWriter, _ *http.Request) {
	type dto struct {
		ID          string `json:"id"`
		Nombre      string `json:"nombre"`
		Tipo        string `json:"tipo"`
		Descripcion string `json:"descripcion"`
		Asunto      string `json:"asunto"`
		Remitente   string `json:"remitente"`
	}
	lista := []dto{}
	for _, p := range plantillas.Todas() {
		lista = append(lista, dto{p.ID, p.Nombre, p.Tipo, p.Descripcion, p.Asunto, p.RemitenteNombre + " <" + p.RemitenteCorreo + ">"})
	}
	web.JSON(w, http.StatusOK, lista)
}

// ---------- Empleados ----------

func (s *servidor) listarAreas(w http.ResponseWriter, r *http.Request) {
	lista, err := s.repo.listarAreas(r.Context())
	if err != nil {
		errorInterno(w, "listar áreas", err)
		return
	}
	web.JSON(w, http.StatusOK, lista)
}

func (s *servidor) listarEmpleados(w http.ResponseWriter, r *http.Request) {
	lista, err := s.repo.listarEmpleados(r.Context())
	if err != nil {
		errorInterno(w, "listar empleados", err)
		return
	}
	web.JSON(w, http.StatusOK, lista)
}

func (s *servidor) crearEmpleado(w http.ResponseWriter, r *http.Request) {
	var e Empleado
	if err := web.LeerJSON(w, r, &e); err != nil {
		web.Error(w, http.StatusBadRequest, "Los datos enviados no son válidos.")
		return
	}
	e.Nombre, e.Correo, e.Area = strings.TrimSpace(e.Nombre), strings.TrimSpace(e.Correo), strings.TrimSpace(e.Area)
	if len(e.Nombre) < 2 || len(e.Nombre) > 120 {
		web.Error(w, http.StatusUnprocessableEntity, "El nombre debe tener entre 2 y 120 caracteres.")
		return
	}
	if !strings.Contains(e.Correo, "@") || len(e.Correo) > 200 {
		web.Error(w, http.StatusUnprocessableEntity, "Escribe un correo válido.")
		return
	}
	if len(e.Area) < 2 || len(e.Area) > 80 {
		web.Error(w, http.StatusUnprocessableEntity, "El área debe tener entre 2 y 80 caracteres.")
		return
	}
	creado, err := s.repo.crearEmpleado(r.Context(), e)
	if errors.Is(err, errCorreoDuplicado) {
		web.Error(w, http.StatusConflict, "Ya existe un empleado con ese correo.")
		return
	}
	if err != nil {
		errorInterno(w, "crear empleado", err)
		return
	}
	web.JSON(w, http.StatusCreated, creado)
}

// ---------- Campañas ----------

type solicitudCampana struct {
	Nombre      string   `json:"nombre"`
	PlantillaID string   `json:"plantilla_id"`
	Areas       []string `json:"areas"`
}

// crearCampana responde de inmediato (201). El envío de correos ocurre después,
// en segundo plano, cuando el servicio "correos" consuma el evento.
func (s *servidor) crearCampana(w http.ResponseWriter, r *http.Request) {
	var sol solicitudCampana
	if err := web.LeerJSON(w, r, &sol); err != nil {
		web.Error(w, http.StatusBadRequest, "Los datos enviados no son válidos.")
		return
	}
	sol.Nombre = strings.TrimSpace(sol.Nombre)
	if len(sol.Nombre) < 3 || len(sol.Nombre) > 120 {
		web.Error(w, http.StatusUnprocessableEntity, "El nombre de la campaña debe tener entre 3 y 120 caracteres.")
		return
	}
	if _, ok := plantillas.Buscar(sol.PlantillaID); !ok {
		web.Error(w, http.StatusUnprocessableEntity, "Elige una plantilla de la lista.")
		return
	}
	usuario := r.Header.Get("X-Usuario") // lo agrega NGINX tras validar el JWT
	if usuario == "" {
		usuario = "desconocido"
	}
	campana, err := s.repo.crearCampana(r.Context(), usuario, sol.Nombre, sol.PlantillaID, sol.Areas)
	if errors.Is(err, errSinDestinatarios) {
		web.Error(w, http.StatusUnprocessableEntity, "No hay empleados en las áreas elegidas. Agrega empleados o cambia la selección.")
		return
	}
	if err != nil {
		errorInterno(w, "crear campaña", err)
		return
	}
	slog.Info("campaña creada", "campana_id", campana.ID, "destinatarios", campana.TotalDestinatarios, "usuario", usuario)
	web.JSON(w, http.StatusCreated, campana)
}

func (s *servidor) listarCampanas(w http.ResponseWriter, r *http.Request) {
	lista, err := s.repo.listarCampanas(r.Context())
	if err != nil {
		errorInterno(w, "listar campañas", err)
		return
	}
	web.JSON(w, http.StatusOK, lista)
}

func (s *servidor) obtenerCampana(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !web.EsUUID(id) {
		web.Error(w, http.StatusNotFound, "No encontramos esa campaña.")
		return
	}
	c, err := s.repo.obtenerCampana(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		web.Error(w, http.StatusNotFound, "No encontramos esa campaña.")
		return
	}
	if err != nil {
		errorInterno(w, "obtener campaña", err)
		return
	}
	web.JSON(w, http.StatusOK, c)
}

// ---------- Rutas públicas ----------

func tokenValido(t string) bool {
	if len(t) != 32 {
		return false
	}
	for _, c := range t {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// interaccion registra el clic o el reporte y redirige a la página que corresponde.
func (s *servidor) interaccion(w http.ResponseWriter, r *http.Request, tipo, destino string) {
	token := r.PathValue("token")
	w.Header().Set("Cache-Control", "no-store")
	if !tokenValido(token) {
		http.Redirect(w, r, "/capacitacion/invalido", http.StatusFound)
		return
	}
	existe, err := s.repo.registrarInteraccion(r.Context(), token, tipo)
	if err != nil {
		slog.Error("registrar interacción", "tipo", tipo, "error", err)
		http.Error(w, "Ocurrió un error. Inténtalo de nuevo.", http.StatusInternalServerError)
		return
	}
	if !existe {
		http.Redirect(w, r, "/capacitacion/invalido", http.StatusFound)
		return
	}
	http.Redirect(w, r, destino+token, http.StatusFound)
}

func (s *servidor) clic(w http.ResponseWriter, r *http.Request) {
	s.interaccion(w, r, "clic", "/capacitacion/")
}

func (s *servidor) reporte(w http.ResponseWriter, r *http.Request) {
	s.interaccion(w, r, "reporte", "/reportado/")
}

func (s *servidor) capacitacion(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if !tokenValido(token) {
		web.Error(w, http.StatusNotFound, "Este enlace no es válido.")
		return
	}
	info, err := s.repo.infoCapacitacion(r.Context(), token)
	if errors.Is(err, pgx.ErrNoRows) {
		web.Error(w, http.StatusNotFound, "Este enlace no es válido.")
		return
	}
	if err != nil {
		errorInterno(w, "info de capacitación", err)
		return
	}
	p, _ := plantillas.Buscar(info.PlantillaID)
	web.JSON(w, http.StatusOK, map[string]any{
		"nombre":    strings.Fields(info.Nombre + " ")[0], // solo el primer nombre
		"clic":      info.Clic,
		"reportado": info.Reportado,
		"plantilla": map[string]any{
			"nombre": p.Nombre, "tipo": p.Tipo, "senales": p.Senales, "consejo": p.Consejo,
		},
	})
}
