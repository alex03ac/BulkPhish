#!/usr/bin/env bash
# Crea el archivo ".env" con claves aleatorias y únicas. Ejecútalo UNA vez, antes de "docker compose up".
# No sobrescribe un .env existente. El archivo queda con permisos 600 y está en .gitignore.
set -euo pipefail
cd "$(dirname "$0")/.."

if [ -e .env ]; then
  echo "Ya existe un archivo .env: no se modifica. Bórralo si quieres generar uno nuevo."
  exit 1
fi

aleatorio() { # aleatorio <bytes>: texto hexadecimal aleatorio
  if command -v openssl >/dev/null 2>&1; then openssl rand -hex "$1"
  else head -c "$1" /dev/urandom | od -An -tx1 | tr -d ' \n'; fi
}
clave_humana() { # 20 caracteres alfanuméricos (sin símbolos que rompan archivos o URLs)
  # (|| true: head cierra la tubería al llegar a 20 caracteres y eso no es un error)
  LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 20 || true
}

ADMIN_CLAVE_GENERADA="$(clave_humana)"
[ "${#ADMIN_CLAVE_GENERADA}" -eq 20 ] || { echo "No se pudo generar una clave aleatoria."; exit 1; }
umask 077
cat > .env <<ENV
# Generado por scripts/generar-env.sh. NO subir a Git.
JWT_SECRETO=$(aleatorio 32)
ADMIN_CORREO=admin@bulkphish.test
ADMIN_NOMBRE=Administrador
ADMIN_CLAVE=${ADMIN_CLAVE_GENERADA}
CLAVE_BD_ADMIN=$(aleatorio 24)
CLAVE_BD_AUTENTICACION=$(aleatorio 24)
CLAVE_BD_CAMPANAS=$(aleatorio 24)
CLAVE_BD_RESULTADOS=$(aleatorio 24)
RABBIT_USUARIO=bulkphish
RABBIT_CLAVE=$(aleatorio 24)
GRAFANA_USUARIO=admin
GRAFANA_CLAVE=$(aleatorio 24)
BIND_GATEWAY=127.0.0.1
URL_PUBLICA=http://localhost:8080
ENV
chmod 600 .env

# Las contraseñas de PostgreSQL solo se aplican cuando el volumen de datos se crea por primera vez.
if command -v docker >/dev/null 2>&1 && docker volume ls -q 2>/dev/null | grep -q '^bulkphish_datos_postgres$'; then
  echo "⚠  Ya existe un volumen de datos de una ejecución anterior (bulkphish_datos_postgres)."
  echo "   Conserva las contraseñas viejas y los servicios no podrán conectarse."
  echo "   Antes de levantar, ejecuta:  docker compose down -v   (borra los datos de prueba)"
  echo
fi
echo "Archivo .env creado con claves aleatorias."
echo
echo "  Usuario administrador : admin@bulkphish.test"
echo "  Contraseña            : ${ADMIN_CLAVE_GENERADA}"
echo
echo "Guarda esa contraseña en un lugar seguro (también está dentro de .env)."
echo "Ahora ejecuta:  docker compose up --build"
