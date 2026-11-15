// Cliente HTTP: agrega el token a cada petición y traduce los errores a mensajes legibles.

const CLAVE_SESION = 'bulkphish.sesion';

export class ErrorApi extends Error {
  constructor(public estado: number, mensaje: string) {
    super(mensaje);
  }
}

export interface Sesion {
  token: string;
  usuario: { nombre: string; correo: string };
  expira_en: string;
}

// La sesión vive en sessionStorage: se borra al cerrar la pestaña (menos exposición que localStorage).
export function leerSesion(): Sesion | null {
  try {
    const crudo = sessionStorage.getItem(CLAVE_SESION);
    if (!crudo) return null;
    const sesion = JSON.parse(crudo) as Sesion;
    return new Date(sesion.expira_en).getTime() > Date.now() ? sesion : null;
  } catch {
    return null;
  }
}
export function guardarSesion(sesion: Sesion) {
  sessionStorage.setItem(CLAVE_SESION, JSON.stringify(sesion));
}
export function borrarSesion() {
  sessionStorage.removeItem(CLAVE_SESION);
}

export const EVENTO_SESION_EXPIRADA = 'bulkphish:sesion-expirada';

export async function llamar<T>(ruta: string, opciones: { metodo?: string; cuerpo?: unknown; publica?: boolean } = {}): Promise<T> {
  const cabeceras: Record<string, string> = { Accept: 'application/json' };
  if (opciones.cuerpo !== undefined) cabeceras['Content-Type'] = 'application/json';
  const sesion = leerSesion();
  if (sesion && !opciones.publica) cabeceras['Authorization'] = `Bearer ${sesion.token}`;

  let respuesta: Response;
  try {
    respuesta = await fetch(ruta, {
      method: opciones.metodo ?? 'GET',
      headers: cabeceras,
      body: opciones.cuerpo !== undefined ? JSON.stringify(opciones.cuerpo) : undefined,
    });
  } catch {
    throw new ErrorApi(0, 'No se pudo conectar con el servidor. Revisa tu conexión e inténtalo de nuevo.');
  }

  if (respuesta.status === 401 && !opciones.publica && ruta !== '/api/auth/login') {
    borrarSesion();
    window.dispatchEvent(new Event(EVENTO_SESION_EXPIRADA));
    throw new ErrorApi(401, 'La sesión expiró. Inicia sesión de nuevo.');
  }
  if (!respuesta.ok) {
    let mensaje = 'Ocurrió un error inesperado. Inténtalo de nuevo.';
    try {
      const cuerpo = await respuesta.json();
      if (cuerpo?.error) mensaje = cuerpo.error;
    } catch {
      /* respuesta sin JSON */
    }
    throw new ErrorApi(respuesta.status, mensaje);
  }
  return (await respuesta.json()) as T;
}
