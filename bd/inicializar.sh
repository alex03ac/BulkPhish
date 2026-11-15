#!/bin/sh
# Se ejecuta UNA sola vez, al crear el volumen de PostgreSQL por primera vez.
# Crea un esquema y un rol por servicio: cada servicio solo puede tocar sus propias tablas.
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<SQL
CREATE ROLE rol_autenticacion LOGIN PASSWORD '${CLAVE_BD_AUTENTICACION}';
CREATE ROLE rol_campanas      LOGIN PASSWORD '${CLAVE_BD_CAMPANAS}';
CREATE ROLE rol_resultados    LOGIN PASSWORD '${CLAVE_BD_RESULTADOS}';

CREATE SCHEMA autenticacion AUTHORIZATION rol_autenticacion;
CREATE SCHEMA campanas      AUTHORIZATION rol_campanas;
CREATE SCHEMA resultados    AUTHORIZATION rol_resultados;

ALTER ROLE rol_autenticacion SET search_path = autenticacion;
ALTER ROLE rol_campanas      SET search_path = campanas;
ALTER ROLE rol_resultados    SET search_path = resultados;

-- Mínimo privilegio: nadie entra a la base ni usa el esquema público salvo quien se autoriza expresamente.
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON DATABASE ${POSTGRES_DB} FROM PUBLIC;
GRANT CONNECT ON DATABASE ${POSTGRES_DB} TO rol_autenticacion, rol_campanas, rol_resultados;
SQL

psql -v ON_ERROR_STOP=1 --username rol_autenticacion --dbname "$POSTGRES_DB" -f /esquemas/autenticacion.sql
psql -v ON_ERROR_STOP=1 --username rol_campanas      --dbname "$POSTGRES_DB" -f /esquemas/campanas.sql
psql -v ON_ERROR_STOP=1 --username rol_campanas      --dbname "$POSTGRES_DB" -f /esquemas/campanas_datos.sql
psql -v ON_ERROR_STOP=1 --username rol_resultados    --dbname "$POSTGRES_DB" -f /esquemas/resultados.sql
