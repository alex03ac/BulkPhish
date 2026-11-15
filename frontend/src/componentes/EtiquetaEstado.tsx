import { EstadoDestinatario } from '../api/tipos';

const TEXTOS: Record<EstadoDestinatario, string> = {
  pendiente: 'Sin enviar',
  enviado: 'Sin acción',
  clic: 'Hizo clic',
  reportado: 'Reportó',
};

export function EtiquetaEstado({ estado }: { estado: EstadoDestinatario }) {
  return <span className={`etiqueta ${estado}`}>{TEXTOS[estado]}</span>;
}
