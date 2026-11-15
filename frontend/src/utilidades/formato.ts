const formatoPorcentaje = new Intl.NumberFormat('es-PE', { style: 'percent', maximumFractionDigits: 1 });
const formatoNumero = new Intl.NumberFormat('es-PE');
const formatoFecha = new Intl.DateTimeFormat('es-PE', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
const formatoHora = new Intl.DateTimeFormat('es-PE', { hour: '2-digit', minute: '2-digit', second: '2-digit' });

export const porcentaje = (v: number) => formatoPorcentaje.format(v).replace(/\s/g, '\u00a0');
export const numero = (v: number) => formatoNumero.format(v);
export const fechaHora = (iso: string) => formatoFecha.format(new Date(iso));
export const hora = (iso: string) => formatoHora.format(new Date(iso));

export function haceCuanto(iso: string): string {
  const segundos = Math.round((Date.now() - new Date(iso).getTime()) / 1000);
  if (segundos < 45) return 'hace unos segundos';
  const minutos = Math.round(segundos / 60);
  if (minutos < 60) return `hace ${minutos} min`;
  const horas = Math.round(minutos / 60);
  if (horas < 24) return `hace ${horas} h`;
  return `hace ${Math.round(horas / 24)} d`;
}

export const NOMBRES_PLANTILLA: Record<string, string> = {
  'restablecer-clave': 'Restablecimiento de contraseña',
  'factura-pendiente': 'Factura pendiente de pago',
  'entrega-paquete': 'Paquete retenido en aduana',
  'beneficio-rrhh': 'Beneficio de Recursos Humanos',
};
export const nombrePlantilla = (id: string) => NOMBRES_PLANTILLA[id] ?? id;
