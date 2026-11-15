import { Destinatario } from '../api/tipos';

/** Clics acumulados desde el inicio de la campaña, dibujados como SVG (sin librerías). */
export function GraficoClics({ destinatarios, inicio }: { destinatarios: Destinatario[]; inicio: string }) {
  const t0 = new Date(inicio).getTime();
  const clics = destinatarios
    .filter((d) => d.clic_en)
    .map((d) => Math.max(0, (new Date(d.clic_en as string).getTime() - t0) / 60000))
    .sort((a, b) => a - b);

  if (clics.length === 0) {
    return <div className="vacio" style={{ padding: '24px 0' }}>Todavía no hay clics. Cuando alguien abra el enlace, aparecerá aquí.</div>;
  }

  const ancho = 520, alto = 190, mI = 34, mD = 12, mS = 12, mB = 28;
  const maxX = Math.max(1, clics[clics.length - 1] * 1.05);
  const maxY = clics.length;
  const x = (v: number) => mI + (v / maxX) * (ancho - mI - mD);
  const y = (v: number) => alto - mB - (v / maxY) * (alto - mB - mS);

  let trazo = `M ${x(0)} ${y(0)}`;
  clics.forEach((t, i) => {
    trazo += ` L ${x(t)} ${y(i)} L ${x(t)} ${y(i + 1)}`;
  });
  trazo += ` L ${x(maxX)} ${y(maxY)}`;

  const marcasY = [0, Math.ceil(maxY / 2), maxY];
  const marcasX = [0, maxX / 2, maxX];
  return (
    <div className="grafico">
      <svg viewBox={`0 0 ${ancho} ${alto}`} role="img" aria-label={`Clics acumulados: ${clics.length} en ${Math.ceil(maxX)} minutos`}>
        {marcasY.map((v) => (
          <g key={v}>
            <line x1={mI} x2={ancho - mD} y1={y(v)} y2={y(v)} stroke="#e3e9ed" />
            <text x={mI - 8} y={y(v) + 4} textAnchor="end" fontSize="11" fill="#52656f">{v}</text>
          </g>
        ))}
        {marcasX.map((v, i) => (
          <text key={i} x={x(v)} y={alto - 8} textAnchor={i === 0 ? 'start' : i === 2 ? 'end' : 'middle'} fontSize="11" fill="#52656f">
            {v < 1 ? '0 min' : `${Math.round(v)} min`}
          </text>
        ))}
        <path d={trazo} fill="none" stroke="#d64550" strokeWidth="2.5" strokeLinejoin="round" />
      </svg>
    </div>
  );
}
