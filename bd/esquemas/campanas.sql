-- Dueño: servicio "campanas".
CREATE TABLE empleados (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre     text NOT NULL,
    correo     text NOT NULL UNIQUE,
    area       text NOT NULL,
    creado_en  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_empleados_area ON empleados (area);

CREATE TABLE campanas (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nombre        text NOT NULL,
    plantilla_id  text NOT NULL,
    creada_por    text NOT NULL,
    creada_en     timestamptz NOT NULL DEFAULT now()
);

-- Un destinatario = un empleado dentro de una campaña. El token va en los enlaces del correo.
CREATE TABLE destinatarios (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campana_id    uuid NOT NULL REFERENCES campanas (id) ON DELETE CASCADE,
    empleado_id   uuid NOT NULL REFERENCES empleados (id),
    token         text NOT NULL UNIQUE,
    clic_en       timestamptz,
    reportado_en  timestamptz
);
CREATE INDEX idx_destinatarios_campana ON destinatarios (campana_id);

-- Bandeja de salida (patrón outbox): eventos pendientes de publicar en RabbitMQ.
CREATE TABLE bandeja_salida (
    id            uuid PRIMARY KEY,
    clave_ruta    text NOT NULL,
    sobre         jsonb NOT NULL,
    creado_en     timestamptz NOT NULL DEFAULT now(),
    publicado_en  timestamptz
);
CREATE INDEX idx_bandeja_pendientes ON bandeja_salida (creado_en) WHERE publicado_en IS NULL;
