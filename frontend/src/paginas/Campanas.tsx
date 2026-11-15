import { Link } from 'react-router-dom';
import { llamar } from '../api/cliente';
import { ResumenCampana } from '../api/tipos';
import { ErrorCarga, Esqueleto, Vacio } from '../componentes/Estados';
import { useConsulta } from '../hooks/useConsulta';
import { fechaHora, nombrePlantilla, porcentaje } from '../utilidades/formato';

export function Campanas() {
  const { datos, error, cargando, recargar } = useConsulta(() => llamar<ResumenCampana[]>('/api/resultados/campanas'), [], 8000);
  return (
    <>
      <header className="cabecera">
        <div>
          <h1>Campañas</h1>
          <p>Cada campaña envía un correo simulado a un grupo de empleados y registra quién hace clic o lo reporta.</p>
        </div>
        <Link to="/campanas/nueva" className="boton">Nueva campaña</Link>
      </header>
      {cargando && <Esqueleto alto={260} />}
      {error && !datos && <ErrorCarga error={error} alReintentar={recargar} />}
      {datos && (
        <section className="tarjeta" aria-label="Lista de campañas">
          {datos.length === 0 ? (
            <Vacio titulo="Todavía no hay campañas">
              <p>Elige un tipo de ataque y un grupo de empleados para lanzar la primera.</p>
              <Link to="/campanas/nueva" className="boton">Crear campaña</Link>
            </Vacio>
          ) : (
            <div className="envoltura-tabla">
              <table>
                <thead><tr><th>Campaña</th><th>Tipo de ataque</th><th>Creada</th><th className="num">Enviados</th><th className="num">Clics</th><th className="num">Reportes</th></tr></thead>
                <tbody>
                  {datos.map((c) => (
                    <tr key={c.id}>
                      <td><Link className="celda-principal" to={`/campanas/${c.id}`}>{c.nombre}</Link></td>
                      <td className="suave">{nombrePlantilla(c.plantilla_id)}</td>
                      <td className="suave">{fechaHora(c.creada_en)}</td>
                      <td className="num">{c.enviados}/{c.total}</td>
                      <td className="num">{c.enviados ? `${c.clics} · ${porcentaje(c.tasa_clic)}` : '—'}</td>
                      <td className="num">{c.enviados ? `${c.reportados} · ${porcentaje(c.tasa_reporte)}` : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>
      )}
    </>
  );
}
