import { Grupo } from '../api/tipos';
import { porcentaje } from '../utilidades/formato';

/** Barras horizontales con la tasa de clic por grupo (área o tipo de ataque). */
export function BarrasGrupo({ grupos, etiquetar = (c: string) => c }: { grupos: Grupo[]; etiquetar?: (clave: string) => string }) {
  const ordenados = [...grupos].sort((a, b) => b.tasa_clic - a.tasa_clic);
  return (
    <div className="barras" role="list">
      {ordenados.map((g) => (
        <div className="barra-fila" role="listitem" key={g.clave}>
          <span className="barra-nombre" title={etiquetar(g.clave)}>
            {etiquetar(g.clave)}
          </span>
          <div className="barra-pista" aria-hidden="true">
            <div className="barra-relleno" style={{ width: `${Math.min(100, g.tasa_clic * 100)}%` }} />
          </div>
          <span className="barra-valor">{g.enviados > 0 ? porcentaje(g.tasa_clic) : '—'}</span>
          <span className="barra-detalle">
            {g.enviados > 0 ? `${g.clics} de ${g.enviados} hicieron clic · ${g.reportados} reportaron` : 'Aún sin correos enviados'}
          </span>
        </div>
      ))}
    </div>
  );
}
