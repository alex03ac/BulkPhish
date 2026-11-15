-- Dueño: servicio "autenticacion".
CREATE TABLE administradores (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    correo      text NOT NULL UNIQUE,
    nombre      text NOT NULL,
    clave_hash  text NOT NULL,              -- bcrypt; nunca se guarda la contraseña
    creado_en   timestamptz NOT NULL DEFAULT now()
);
