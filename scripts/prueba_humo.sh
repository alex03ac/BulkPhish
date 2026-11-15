#!/usr/bin/env bash
# Prueba de humo de extremo a extremo: recorre el flujo completo contra el sistema levantado.
#   login → crear campaña → llegan correos → clic → reporte → resultados → seguridad básica
# Uso: ./scripts/prueba_humo.sh        (GATEWAY y MAILPIT se pueden cambiar por variables)
set -uo pipefail

# Lee las credenciales del .env (si existe) sin exportar nada más.
if [ -f "$(dirname "$0")/../.env" ]; then
  ADMIN_CORREO=$(grep -E '^ADMIN_CORREO=' "$(dirname "$0")/../.env" | cut -d= -f2-)
  ADMIN_CLAVE=$(grep -E '^ADMIN_CLAVE=' "$(dirname "$0")/../.env" | cut -d= -f2-)
fi
G=${GATEWAY:-http://localhost:8080}
M=${MAILPIT:-http://localhost:8025}
CORREO=${ADMIN_CORREO:-admin@bulkphish.test}
CLAVE=${ADMIN_CLAVE:?Falta ADMIN_CLAVE (ejecuta scripts/generar-env.sh)}
fallos=0

ok()     { echo "  ✔ $1"; }
fallar() { echo "  ✘ $1"; fallos=$((fallos + 1)); }
cod()    { curl -s -o /dev/null -w '%{http_code}' "$@"; }
campo()  { sed -n "s/.*\"$1\":\"\([^\"]*\)\".*/\1/p" | head -1; }
esperar() { # esperar <segundos> <comando...>: reintenta hasta que el comando tenga éxito
  local limite=$1; shift
  for _ in $(seq 1 "$limite"); do "$@" >/dev/null 2>&1 && return 0; sleep 1; done
  return 1
}

echo "1. Gateway y seguridad básica"
[ "$(cod "$G/estado")" = "200" ] && ok "el gateway responde" || fallar "el gateway no responde en $G"
[ "$(cod "$G/api/campanas")" = "401" ] && ok "las rutas protegidas exigen sesión (401 sin token)" || fallar "una ruta protegida respondió sin token"
[ "$(cod -H 'Authorization: Bearer falso.falso.falso' "$G/api/empleados")" = "401" ] && ok "un token falso es rechazado" || fallar "se aceptó un token falso"
curl -sI "$G/" | grep -qi '^x-frame-options: deny' && ok "cabeceras de seguridad presentes" || fallar "faltan cabeceras de seguridad"

echo "2. Inicio de sesión"
TOKEN=$(curl -s -X POST "$G/api/auth/login" -d "{\"correo\":\"$CORREO\",\"clave\":\"$CLAVE\"}" | campo token)
[ -n "$TOKEN" ] && ok "login correcto, JWT recibido" || { fallar "no se pudo iniciar sesión"; echo "Resultado: $fallos fallo(s)"; exit 1; }
AUTH=(-H "Authorization: Bearer $TOKEN")

echo "3. Crear campaña (síncrono)"
ANTES=$(curl -s "$M/api/v1/messages?limit=1" | sed -n 's/.*"total":\([0-9]*\).*/\1/p' | head -1); ANTES=${ANTES:-0}
RESP=$(curl -s -X POST "$G/api/campanas" "${AUTH[@]}" -d '{"nombre":"Prueba de humo","plantilla_id":"restablecer-clave","areas":["Finanzas"]}')
ID=$(echo "$RESP" | campo id)
[ -n "$ID" ] && ok "campaña creada ($ID)" || { fallar "no se creó la campaña: $RESP"; echo "Resultado: $fallos fallo(s)"; exit 1; }

echo "4. Envío de correos (asíncrono)"
hay_correos() { local n; n=$(curl -s "$M/api/v1/messages?limit=1" | sed -n 's/.*"total":\([0-9]*\).*/\1/p' | head -1); [ "${n:-0}" -gt "$ANTES" ]; }
esperar 40 hay_correos && ok "los correos llegaron al buzón de pruebas" || fallar "no llegaron correos a Mailpit"

echo "5. Clic y reporte desde el correo (rutas públicas)"
MSG=$(curl -s "$M/api/v1/messages?limit=1" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
CUERPO=$(curl -s "$M/api/v1/message/$MSG")
TK=$(echo "$CUERPO" | grep -o '/t/[a-f0-9]\{32\}' | head -1 | cut -d/ -f3)
TR=$(echo "$CUERPO" | grep -o '/r/[a-f0-9]\{32\}' | head -1 | cut -d/ -f3)
[ -n "$TK" ] && [ -n "$TR" ] && ok "el correo trae enlace de clic y de reporte" || fallar "no se encontraron los enlaces en el correo"
[ "$(cod "$G/t/$TK")" = "302" ] && ok "el clic redirige a la capacitación (302)" || fallar "el clic no respondió 302"
[ "$(cod "$G/t/$TK")" = "302" ] && ok "repetir el clic es inofensivo" || fallar "el segundo clic falló"
[ "$(cod "$G/r/$TR")" = "302" ] && ok "el reporte redirige (302)" || fallar "el reporte no respondió 302"

echo "6. Resultados"
hay_clic() { curl -s "${AUTH[@]}" "$G/api/resultados/campanas/$ID" | grep -q '"clics":[1-9]'; }
esperar 30 hay_clic && ok "el clic aparece en los resultados" || fallar "el clic no llegó a resultados"
curl -s "${AUTH[@]}" "$G/api/resultados/resumen" | grep -q '"total_campanas"' && ok "el resumen del panel responde" || fallar "falló el resumen"

echo "7. Protección contra fuerza bruta"
BLOQUEADO=0
for _ in 1 2 3 4 5 6 7 8; do
  [ "$(cod -X POST "$G/api/auth/login" -d '{"correo":"intruso@bulkphish.test","clave":"incorrecta"}')" = "429" ] && BLOQUEADO=1
done
[ "$BLOQUEADO" = "1" ] && ok "tras varios intentos fallidos se bloquea (429)" || fallar "no se bloqueó la fuerza bruta"

echo
if [ "$fallos" -eq 0 ]; then echo "Resultado: TODO CORRECTO ✔"; else echo "Resultado: $fallos fallo(s) ✘"; exit 1; fi
