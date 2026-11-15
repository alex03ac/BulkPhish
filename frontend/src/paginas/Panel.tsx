import { Link } from 'react-router-dom';
import { llamar } from '../api/cliente';
import { ResumenGlobal } from '../api/tipos';
import { BarrasGrupo } from '../componentes/BarrasGrupo';
import { ErrorCarga, Esqueleto, Vacio } from '../componentes/Estados';
import { Kpi } from '../componentes/Kpi';
import { useConsulta } from '../hooks/useConsulta';
import { haceCuanto, nombrePlantilla, numero, porcentaje } from '../utilidades/formato';

export function Panel() {
  const { datos, error, cargando, recargar } = useConsulta(() => llamar<ResumenGlobal>('/api/resultados/resumen'), [], 10000);

  return (
    <>
      <header className="cabecera">
        <div>
          <h1>Panel</h1>
          <p>Cómo responde tu organización a los simulacros de phishing.</p>
        </div>
        <Link to="/campanas/nueva" className="boton">Nueva campaña</Link>
      </header>

      {cargando && <div className="rejilla kpis"><Esqueleto alto={96} /><Esqueleto alto={96} /><Esqueleto alto={96} /><Esqueleto alto={96} /></div>}
      {error && !datos && <ErrorCarga error={error} alReintentar={recargar} />}

      {datos && datos.total_campanas === 0 && (
        <div className="tarjeta">
          <Vacio titulo="Aún no hay campañas">
            <p>Lanza tu primera campaña de práctica para empezar a medir cómo responde cada equipo.</p>
            <Link to="/campanas/nueva" className="boton">Crear la primera campaña</Link>
          </Vacio>
        </div>
      )}

      {datos && datos.total_campanas > 0 && (
        <>
          <div className="rejilla kpis">
            <Kpi etiqueta="Campañas" valor={numero(datos.total_campanas)} nota="realizadas hasta hoy" />
            <Kpi etiqueta="Correos enviados" valor={numero(datos.global.enviados)} nota={`de ${numero(datos.global.total)} previstos`} />
            <Kpi etiqueta="Tasa de clic" valor={porcentaje(datos.global.tasa_clic)} nota={`${numero(datos.global.clics)} personas cayeron`} tono="clic" />
            <Kpi etiqueta="Tasa de reporte" valor={porcentaje(datos.global.tasa_reporte)} nota={`${numero(datos.global.reportados)} lo reportaron`} tono="reporte" />
          </div>

          <div className="rejilla dos" style={{ marginBottom: 18 }}>
            <section className="tarjeta" aria-labelledby="t-area">
              <div className="tarjeta-titulo"><h2 id="t-area">Clics por área</h2><small>Más alto = más riesgo</small></div>
              <BarrasGrupo grupos={datos.por_area} />
            </section>
            <section className="tarjeta" aria-labelledby="t-tipo">
              <div className="tarjeta-titulo"><h2 id="t-tipo">Clics por tipo de ataque</h2><small>Qué engaña más</small></div>
              <BarrasGrupo grupos={datos.por_plantilla} etiquetar={nombrePlantilla} />
            </section>
          </div>

          <section className="tarjeta" aria-labelledby="t-recientes">
            <div className="tarjeta-titulo"><h2 id="t-recientes">Campañas recientes</h2><Link to="/campanas">Ver todas</Link></div>
            <div className="envoltura-tabla">
              <table>
                <thead><tr><th>Campaña</th><th>Tipo de ataque</th><th className="num">Enviados</th><th className="num">Clics</th><th className="num">Reportes</th><th>Creada</th></tr></thead>
                <tbody>
                  {datos.recientes.map((c) => (
                    <tr key={c.id}>
                      <td><Link className="celda-principal" to={`/campanas/${c.id}`}>{c.nombre}</Link></td>
                      <td className="suave">{nombrePlantilla(c.plantilla_id)}</td>
                      <td className="num">{c.enviados}/{c.total}</td>
                      <td className="num">{c.enviados ? porcentaje(c.tasa_clic) : '—'}</td>
                      <td className="num">{c.enviados ? porcentaje(c.tasa_reporte) : '—'}</td>
                      <td className="suave">{haceCuanto(c.creada_en)}</td>
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
