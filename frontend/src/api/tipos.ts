// Tipos que describen las respuestas de la API (coinciden con los JSON de los servicios Go).

export interface Estadisticas {
  total: number;
  enviados: number;
  clics: number;
  reportados: number;
  tasa_clic: number; // 0 a 1
  tasa_reporte: number; // 0 a 1
}

export interface ResumenCampana extends Estadisticas {
  id: string;
  nombre: string;
  plantilla_id: string;
  creada_en: string;
}

export interface Grupo extends Estadisticas {
  clave: string;
}

export interface ResumenGlobal {
  total_campanas: number;
  global: Estadisticas;
  por_area: Grupo[];
  por_plantilla: Grupo[];
  recientes: ResumenCampana[];
}

export type EstadoDestinatario = 'pendiente' | 'enviado' | 'clic' | 'reportado';

export interface Destinatario {
  nombre: string;
  correo: string;
  area: string;
  estado: EstadoDestinatario;
  enviado_en: string | null;
  clic_en: string | null;
  reportado_en: string | null;
}

export interface DetalleResultados {
  campana: ResumenCampana;
  por_area: Grupo[];
  destinatarios: Destinatario[];
}

export interface CampanaBase {
  id: string;
  nombre: string;
  plantilla_id: string;
  creada_por: string;
  creada_en: string;
  total_destinatarios: number;
  areas: string[];
}

export interface Plantilla {
  id: string;
  nombre: string;
  tipo: string;
  descripcion: string;
  asunto: string;
  remitente: string;
}

export interface Empleado {
  id: string;
  nombre: string;
  correo: string;
  area: string;
}

export interface AreaConteo {
  area: string;
  empleados: number;
}

export interface Capacitacion {
  nombre: string;
  clic: boolean;
  reportado: boolean;
  plantilla: { nombre: string; tipo: string; senales: string[]; consejo: string };
}

export interface Usuario {
  nombre: string;
  correo: string;
}
