import { ArrowLeft } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { llamar } from '../api/cliente';
import { CampanaBase, DetalleResultados } from '../api/tipos';
import { BarrasGrupo } from '../componentes/BarrasGrupo';
import { ErrorCarga, Esqueleto } from '../componentes/Estados';
import { EtiquetaEstado } from '../componentes/EtiquetaEstado';
import { GraficoClics } from '../componentes/GraficoLinea';
import { Kpi } from '../componentes/Kpi';
import { Mosaico } from '../componentes/Mosaico';
import { useConsulta } from '../hooks/useConsulta';
import { fechaHora, hora, nombrePlantilla, numero, porcentaje } from '../utilidades/formato';

export function DetalleCampana() {
  const { id = '' } = useParams();
  const base = useConsulta(() => llamar<CampanaBase>(`/api/campanas/${id}`), [id]);
  // Se consulta cada 3 s: así se ven los clics a medida que ocurren.
  const resultados = useConsulta(() => llamar<DetalleResultados>(`/api/resultados/campanas/${id}`), [id], 3000);

  if (base.error?.estado === 404) {
    return (
      <div className="tarjeta"><div className="vacio"><h3>No encontramos esa campaña</h3><Link to="/campanas" className="boton secundario">Volver a campañas</Link></div></div>
    );
  }

  const detalle = resultados.datos;
  const preparando = !detalle && (resultados.error?.estado === 404 || resultados.cargando);
  const c = detalle?.campana;
  const enCurso = !!c && c.enviados < c.total;

  return (
    <>
      <header className="cabecera">
        <div>
          <Link to="/campanas" className="migas"><ArrowLeft size={14} aria-hidden="true" /> Campañas</Link>
          <h1>{base.datos?.nombre ?? 'Campaña'}</h1>
          {base.datos && (
            <p>
              {nombrePlantilla(base.datos.plantilla_id)} · creada el {fechaHora(base.datos.creada_en)} por {base.datos.creada_por}
            </p>
          )}
          {enCurso && c && (
            <>
              <p className="aviso-vivo" style={{ marginTop: 10 }}><i aria-hidden="true" />Enviando correos: {c.enviados} de {c.total}</p>
              <div className="progreso" role="progressbar" aria-valuemin={0} aria-valuemax={c.total} aria-valuenow={c.enviados}><div style={{ width: `${(c.enviados / c.total) * 100}%` }} /></div>
            </>
          )}
          {!enCurso && c && <p className="aviso-vivo" style={{ marginTop: 10 }}><i aria-hidden="true" />Resultados en vivo · se actualizan cada 3 segundos</p>}
        </div>
      </header>

      {preparando && (
        <div className="tarjeta" role="status">
          <div className="vacio">
            <h3>Preparando la campaña…</h3>
            <p>Los correos están en cola. En unos segundos verás los resultados aquí.</p>
          </div>
          <Esqueleto alto={80} />
        </div>
      )}
      {resultados.error && !detalle && resultados.error.estado !== 404 && <ErrorCarga error={resultados.error} alReintentar={resultados.recargar} />}

      {detalle && c && (
        <>
          <div className="rejilla kpis">
            <Kpi etiqueta="Correos enviados" valor={`${c.enviados}/${c.total}`} nota={enCurso ? 'envío en curso' : 'envío completo'} />
            <Kpi etiqueta="Tasa de clic" valor={c.enviados ? porcentaje(c.tasa_clic) : '—'} nota={`${numero(c.clics)} personas`} tono="clic" />
            <Kpi etiqueta="Tasa de reporte" valor={c.enviados ? porcentaje(c.tasa_reporte) : '—'} nota={`${numero(c.reportados)} personas`} tono="reporte" />
            <Kpi etiqueta="Sin acción" valor={numero(Math.max(0, c.enviados - c.clics - c.reportados))} nota="recibieron y no interactuaron" />
          </div>

          <section className="tarjeta" aria-labelledby="t-mosaico" style={{ marginBottom: 18 }}>
            <div className="tarjeta-titulo"><h2 id="t-mosaico">Quién hizo qué</h2><small>Pasa el cursor sobre una celda para ver el nombre</small></div>
            <Mosaico destinatarios={detalle.destinatarios} />
          </section>

          <div className="rejilla dos" style={{ marginBottom: 18 }}>
            <section className="tarjeta" aria-labelledby="t-area">
              <div className="tarjeta-titulo"><h2 id="t-area">Clics por área</h2></div>
              <BarrasGrupo grupos={detalle.por_area} />
            </section>
            <section className="tarjeta" aria-labelledby="t-evol">
              <div className="tarjeta-titulo"><h2 id="t-evol">Clics acumulados</h2><small>desde el lanzamiento</small></div>
              <GraficoClics destinatarios={detalle.destinatarios} inicio={c.creada_en} />
            </section>
          </div>

          <section className="tarjeta" aria-labelledby="t-personas">
            <div className="tarjeta-titulo"><h2 id="t-personas">Destinatarios</h2><small>{detalle.destinatarios.length} personas</small></div>
            <div className="envoltura-tabla">
              <table>
                <thead><tr><th>Nombre</th><th>Área</th><th>Estado</th><th>Hora del clic</th></tr></thead>
                <tbody>
                  {detalle.destinatarios.map((d) => (
                    <tr key={d.correo}>
                      <td><strong>{d.nombre}</strong><div className="suave">{d.correo}</div></td>
                      <td className="suave">{d.area}</td>
                      <td><EtiquetaEstado estado={d.estado} /></td>
                      <td className="suave">{d.clic_en ? hora(d.clic_en) : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        </>
      )}
    </>
  );
}
