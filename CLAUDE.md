# CLAUDE.md

Guía para trabajar en **BulkPhish**. Lee también `README.md` y `doc/ARQUITECTURA.md` para el detalle.

## Qué es

Plataforma web para **simular campañas de phishing** y capacitar empleados (uso educativo/defensivo autorizado, tipo GoPhish). Los correos nunca salen a Internet: van a Mailpit (buzón de pruebas). Proyecto del curso **Desarrollo Web Avanzado**.

## Convención clave: TODO en español

Nombres de servicios, paquetes, carpetas, tipos, variables, comentarios y mensajes de error están **en español**. Mantén ese idioma al añadir código (`autenticacion`, `campanas`, `correos`, `resultados`, `plantillas`, etc.). No introduzcas nombres en inglés.

## Arquitectura (orientada a eventos)

4 microservicios **Go** + gateway **NGINX** + frontend **React**. Los servicios **no se llaman entre sí**: se comunican publicando/consumiendo eventos por **RabbitMQ**. El gateway es la única entrada y valida el JWT antes de rutas protegidas.

| Servicio (`cmd/<x>` + `internal/<x>`) | Responsabilidad |
|---|---|
| `autenticacion` | Login, emisión y validación de JWT, anti fuerza bruta |
| `campanas` | Campañas, empleados, clics, reportes, bandeja de salida; **publica** eventos |
| `correos` | **Consume** eventos y envía correos simulados (SMTP → Mailpit) |
| `resultados` | **Consume** todos los eventos, calcula estadísticas (modelo de lectura) |

Paquetes compartidos: `internal/eventos` (contrato de eventos), `internal/mensajeria` (cliente RabbitMQ), `internal/basedatos`, `internal/web`, `internal/config`, `internal/plantillas` (catálogo de ataques + capacitación).

Flujo: `campanas` publica `campana.creada` → `correos` la consume y envía → publica `correo.enviado` → `resultados` consume todo y agrega estadísticas.

## Comandos

```bash
# Levantar todo el sistema (requiere .env, ver abajo)
docker compose up --build

# Crear .env la primera vez (claves aleatorias; sin esto no arranca)
./scripts/generar-env.sh          # Windows: .\scripts\generar-env.ps1

# Backend Go (desde la raíz)
go vet ./...
go test ./...                     # hay tests en internal/*
go mod tidy                       # go.sum ya está commiteado

# Frontend (requiere el sistema levantado)
cd frontend && npm install && npm start   # localhost:3000, proxy /api al gateway

# Prueba de humo end-to-end (con el sistema levantado)
./scripts/prueba_humo.sh

# Antes de cada git push
./scripts/revisar-secretos.sh
```

Puertos: app en `http://localhost:8080`, Mailpit en `:8025`, RabbitMQ en `:15672`, **Grafana en `:3000`**, **Prometheus en `:9090`**. Todo escucha solo en `127.0.0.1`.

## Observabilidad (Fase 3)

Métricas Prometheus en todos los servicios Go (`internal/metricas`): peticiones y latencia HTTP (vía `web.Middleware`), eventos publicados/procesados (vía `mensajeria`) y correos enviados. Cada servicio expone `/metrics` (lo añade `web.Servir`); `correos` levanta un servidor mínimo solo para eso en el puerto 8084. Config en `observabilidad/`: `prometheus.yml` (scrapea los 4 servicios + RabbitMQ) y Grafana con datasource y dashboard **provisionados** (`BulkPhish — Observabilidad`). Grafana usa `GRAFANA_CLAVE` del `.env`.

## Frontend

React 18 + TypeScript + Webpack 5. Carpetas en `frontend/src/`: `api`, `autenticacion`, `componentes`, `hooks`, `paginas`, `utilidades`. La sesión vive en `sessionStorage` (`src/api/cliente.ts`). **CSP estricta**: si algo no se ve o no carga, revisa la consola del navegador antes de tocar el código.

## Reglas de seguridad (no romper)

- **Nunca** pongas secretos en el código ni en `docker-compose.yml`: todo se lee de `.env` (en `.gitignore`). Si falta una variable, el sistema debe detenerse.
- Contenedores sin root, FS de solo lectura, `cap_drop: ALL`. Si tocas `docker-compose.yml`, mantén el endurecimiento.
- SQL siempre parametrizado; contraseñas con bcrypt; no escribir tokens de enlace en logs.
- Un esquema y un rol de PostgreSQL por servicio (mínimo privilegio).

## Estado del proyecto

Fase 1 (MVP) y Fase 2 (JWT, plantillas, resultados por área, botón reportar) **hechas**. Fase 3: **Observabilidad (Prometheus + Grafana) HECHA** (ver sección arriba). **Pendiente de Fase 3**: CI/CD con GitHub Actions y escaneo de seguridad con OWASP ZAP. Detalle en `../FASES_PENDIENTES.md` (fuera del repo). El sistema ya corre con `docker compose up` en WSL2.
