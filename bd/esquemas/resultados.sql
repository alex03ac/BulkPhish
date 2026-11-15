-- Dueño: servicio "resultados". Es un modelo de lectura construido solo con eventos.
CREATE TABLE campanas (
    campana_id    uuid PRIMARY KEY,
    nombre        text NOT NULL,
    plantilla_id  text NOT NULL,
    creada_en     timestamptz NOT NULL
);

CREATE TABLE destinatarios (
    destinatario_id  uuid PRIMARY KEY,
    campana_id       uuid NOT NULL REFERENCES campanas (campana_id) ON DELETE CASCADE,
    nombre           text NOT NULL,
    correo           text NOT NULL,
    area             text NOT NULL,
    enviado_en       timestamptz,
    clic_en          timestamptz,
    reportado_en     timestamptz
);
CREATE INDEX idx_res_destinatarios_campana ON destinatarios (campana_id);
