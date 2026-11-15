import { useEffect, useRef } from 'react';
import { Destinatario } from '../api/tipos';
import { hora } from '../utilidades/formato';

const TEXTO_ESTADO = { pendiente: 'sin enviar', enviado: 'sin acción', clic: 'hizo clic', reportado: 'reportó' } as const;

function iniciales(nombre: string) {
  const partes = nombre.trim().split(/\s+/);
  return (partes[0][0] + (partes[1]?.[0] ?? '')).toUpperCase();
}

/** Una celda por persona, agrupadas por área y coloreadas según lo que hizo. */
export function Mosaico({ destinatarios }: { destinatarios: Destinatario[] }) {
  // Solo las celdas que cambian de estado DESPUÉS de la primera carga reproducen el latido.
  const previo = useRef(new Map<string, string>());
  useEffect(() => {
    destinatarios.forEach((d) => previo.current.set(d.correo, d.estado));
  }, [destinatarios]);

  const porArea = new Map<string, Destinatario[]>();
  destinatarios.forEach((d) => porArea.set(d.area, [...(porArea.get(d.area) ?? []), d]));

  return (
    <div>
      {[...porArea.entries()].map(([area, personas]) => (
        <div className="mosaico-area" key={area}>
          <div className="mosaico-area-nombre">
            <span>{area}</span>
            <span>{personas.length} {personas.length === 1 ? 'persona' : 'personas'}</span>
          </div>
          <div className="mosaico-celdas">
            {personas.map((p) => (
              <button
                type="button"
                key={`${p.correo}-${p.estado}`}
                className={`celda ${p.estado}${previo.current.has(p.correo) && previo.current.get(p.correo) !== p.estado ? ' nuevo' : ''}`}
                aria-label={`${p.nombre}: ${TEXTO_ESTADO[p.estado]}`}
              >
                {iniciales(p.nombre)}
                <span className="globo" role="tooltip">
                  {p.nombre} · {TEXTO_ESTADO[p.estado]}
                  {p.clic_en ? ` (${hora(p.clic_en)})` : ''}
                </span>
              </button>
            ))}
          </div>
        </div>
      ))}
      <div className="leyenda" aria-hidden="true">
        <span><i style={{ background: 'var(--clic)' }} />Hizo clic</span>
        <span><i style={{ background: 'var(--reporte)' }} />Reportó el correo</span>
        <span><i style={{ background: 'var(--enviado)' }} />Recibió, sin acción</span>
        <span><i style={{ background: 'var(--pendiente)' }} />Aún sin enviar</span>
      </div>
    </div>
  );
}
