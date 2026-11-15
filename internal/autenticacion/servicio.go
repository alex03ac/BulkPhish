// Paquete autenticacion: servicio que identifica a los administradores.
//   - POST /api/auth/login  → entrega un JWT si las credenciales son correctas.
//   - GET  /api/auth/yo     → datos del usuario del token actual.
//   - GET  /validar         → lo consulta NGINX (auth_request) antes de dejar pasar
//     cualquier petición a las rutas protegidas.
package autenticacion

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"bulkphish/internal/basedatos"
	"bulkphish/internal/config"
	"bulkphish/internal/web"
)

type servidor struct {
	limitador *limitador
	bd        *pgxpool.Pool
	emisor    *emisorJWT
	hashFalso []byte // se compara cuando el correo no existe, para que el tiempo de respuesta sea parecido
}

// Ejecutar arranca el servicio y bloquea hasta recibir una señal de apagado.
func Ejecutar() error {
	config.IniciarRegistro("autenticacion")
	puerto := config.Texto("PUERTO", "8081")
	secreto := config.Obligatorio("JWT_SECRETO")
	if len(secreto) < 32 {
		return errors.New("JWT_SECRETO debe tener al menos 32 caracteres")
	}
	adminCorreo := config.Obligatorio("ADMIN_CORREO")
	adminClave := config.Obligatorio("ADMIN_CLAVE")
	if len(adminClave) < 12 || strings.EqualFold(adminClave, adminCorreo) {
		return errors.New("ADMIN_CLAVE debe tener al menos 12 caracteres y no puede ser igual al correo")
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	bd, err := basedatos.Conectar(ctx, config.Obligatorio("BD_URL"))
	if err != nil {
		return err
	}
	defer bd.Close()

	if err := asegurarAdministrador(ctx, bd, adminCorreo, adminClave,
		config.Texto("ADMIN_NOMBRE", "Administrador")); err != nil {
		return err
	}

	falso, _ := bcrypt.GenerateFromPassword([]byte("clave-inexistente"), bcrypt.DefaultCost)
	s := &servidor{
		bd:        bd,
		emisor:    &emisorJWT{secreto: []byte(secreto), duracion: time.Duration(config.Entero("JWT_HORAS", 8)) * time.Hour},
		hashFalso: falso,
		limitador: nuevoLimitador(config.Entero("LOGIN_MAX_FALLOS", 5), time.Duration(config.Entero("LOGIN_VENTANA_MIN", 15))*time.Minute),
	}
	return web.Servir(ctx, puerto, s.rutas())
}

func (s *servidor) rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /salud", web.Salud)
	mux.HandleFunc("POST /api/auth/login", s.iniciarSesion)
	mux.HandleFunc("GET /api/auth/yo", s.quienSoy)
	mux.HandleFunc("GET /validar", s.validar)
	return mux
}

// asegurarAdministrador crea el administrador inicial si todavía no existe.
func asegurarAdministrador(ctx context.Context, bd *pgxpool.Pool, correo, clave, nombre string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(clave), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = bd.Exec(ctx, `INSERT INTO administradores (correo, nombre, clave_hash)
		VALUES (lower($1), $2, $3) ON CONFLICT (correo) DO NOTHING`, correo, nombre, string(hash))
	if err == nil {
		slog.Info("administrador inicial verificado", "correo", correo)
	}
	return err
}

type solicitudLogin struct {
	Correo string `json:"correo"`
	Clave  string `json:"clave"`
}

type usuarioDTO struct {
	Nombre string `json:"nombre"`
	Correo string `json:"correo"`
}

func (s *servidor) iniciarSesion(w http.ResponseWriter, r *http.Request) {
	var sol solicitudLogin
	if err := web.LeerJSON(w, r, &sol); err != nil {
		web.Error(w, http.StatusBadRequest, "Envía el correo y la contraseña.")
		return
	}
	correo := strings.ToLower(strings.TrimSpace(sol.Correo))

	// Anti fuerza bruta: la clave combina correo e IP (la IP real la pone el gateway).
	clave := correo + "|" + ipCliente(r)
	if s.limitador.bloqueado(clave) {
		w.Header().Set("Retry-After", "900")
		web.Error(w, http.StatusTooManyRequests, "Demasiados intentos fallidos. Espera unos minutos antes de volver a intentarlo.")
		return
	}

	var id, nombre, hash string
	err := s.bd.QueryRow(r.Context(),
		`SELECT id::text, nombre, clave_hash FROM administradores WHERE correo = $1`, correo).
		Scan(&id, &nombre, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		_ = bcrypt.CompareHashAndPassword(s.hashFalso, []byte(sol.Clave)) // igualar tiempos
		s.limitador.registrarFallo(clave)
		web.Error(w, http.StatusUnauthorized, "Correo o contraseña incorrectos.")
		return
	}
	if err != nil {
		slog.Error("consulta de administrador", "error", err)
		web.Error(w, http.StatusInternalServerError, "No se pudo iniciar sesión. Inténtalo de nuevo.")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(sol.Clave)) != nil {
		s.limitador.registrarFallo(clave)
		web.Error(w, http.StatusUnauthorized, "Correo o contraseña incorrectos.")
		return
	}

	s.limitador.reiniciar(clave)
	token, expira, err := s.emisor.emitir(id, nombre, correo)
	if err != nil {
		slog.Error("emisión de token", "error", err)
		web.Error(w, http.StatusInternalServerError, "No se pudo iniciar sesión. Inténtalo de nuevo.")
		return
	}
	web.JSON(w, http.StatusOK, map[string]any{
		"token":     token,
		"expira_en": expira,
		"usuario":   usuarioDTO{Nombre: nombre, Correo: correo},
	})
}

// reclamosDe extrae y valida el token del encabezado "Authorization: Bearer ...".
func (s *servidor) reclamosDe(r *http.Request) (*Reclamos, error) {
	cabecera := r.Header.Get("Authorization")
	texto, ok := strings.CutPrefix(cabecera, "Bearer ")
	if !ok || texto == "" {
		return nil, errors.New("falta el token")
	}
	return s.emisor.verificar(texto)
}

func (s *servidor) quienSoy(w http.ResponseWriter, r *http.Request) {
	reclamos, err := s.reclamosDe(r)
	if err != nil {
		web.Error(w, http.StatusUnauthorized, "La sesión expiró. Inicia sesión de nuevo.")
		return
	}
	web.JSON(w, http.StatusOK, usuarioDTO{Nombre: reclamos.Nombre, Correo: reclamos.Correo})
}

// validar responde solo con códigos: 200 (pasa) o 401 (no pasa). NGINX no lee el cuerpo.
// En caso de éxito devuelve el usuario en X-Usuario para que NGINX lo reenvíe al servicio destino.
func (s *servidor) validar(w http.ResponseWriter, r *http.Request) {
	reclamos, err := s.reclamosDe(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.Header().Set("X-Usuario", reclamos.Correo)
	w.WriteHeader(http.StatusOK)
}

// ipCliente toma la IP real que NGINX agrega en X-Real-IP; si no existe usa la del socket.
func ipCliente(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
