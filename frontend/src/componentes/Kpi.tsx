export function Kpi({ etiqueta, valor, nota, tono }: { etiqueta: string; valor: string; nota?: string; tono?: 'clic' | 'reporte' }) {
  return (
    <div className={`tarjeta kpi ${tono ?? ''}`}>
      <div className="kpi-etiqueta">{etiqueta}</div>
      <div className="kpi-valor">{valor}</div>
      {nota && <div className="kpi-nota">{nota}</div>}
    </div>
  );
}
