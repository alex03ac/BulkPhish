import { FormEvent, useState } from 'react';
import { ErrorApi, llamar } from '../api/cliente';
import { Empleado } from '../api/tipos';
import { ErrorCarga, Esqueleto } from '../componentes/Estados';
import { useConsulta } from '../hooks/useConsulta';

export function Empleados() {
  const { datos, error, cargando, recargar } = useConsulta(() => llamar<Empleado[]>('/api/empleados'), []);
  const [nombre, setNombre] = useState('');
  const [correo, setCorreo] = useState('');
  const [area, setArea] = useState('');
  const [mensaje, setMensaje] = useState<{ tipo: 'error' | 'ok'; texto: string } | null>(null);
  const [guardando, setGuardando] = useState(false);

  async function alEnviar(e: FormEvent) {
    e.preventDefault();
    setMensaje(null);
    setGuardando(true);
    try {
      await llamar<Empleado>('/api/empleados', { metodo: 'POST', cuerpo: { nombre, correo, area } });
      setNombre(''); setCorreo(''); setArea('');
      setMensaje({ tipo: 'ok', texto: 'Empleado agregado.' });
      await recargar();
    } catch (err) {
      setMensaje({ tipo: 'error', texto: err instanceof ErrorApi ? err.message : 'No se pudo agregar al empleado.' });
    } finally {
      setGuardando(false);
    }
  }

  const areasExistentes = [...new Set((datos ?? []).map((d) => d.area))];

  return (
    <>
      <header className="cabecera">
        <div>
          <h1>Empleados</h1>
          <p>Las personas que pueden recibir simulacros, agrupadas por área.</p>
        </div>
      </header>
      <div className="rejilla principal-lateral">
        <section className="tarjeta" aria-label="Lista de empleados">
          {cargando && <Esqueleto alto={300} />}
          {error && !datos && <ErrorCarga error={error} alReintentar={recargar} />}
          {datos && (
            <div className="envoltura-tabla">
              <table>
                <thead><tr><th>Nombre</th><th>Correo</th><th>Área</th></tr></thead>
                <tbody>
                  {datos.map((d) => (
                    <tr key={d.id}><td><strong>{d.nombre}</strong></td><td className="suave">{d.correo}</td><td>{d.area}</td></tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </section>

        <aside className="tarjeta" aria-label="Agregar empleado">
          <h2 style={{ marginBottom: 14 }}>Agregar empleado</h2>
          <form onSubmit={alEnviar} noValidate>
            {mensaje && <div className={mensaje.tipo === 'error' ? 'error-formulario' : 'sello bien'} role={mensaje.tipo === 'error' ? 'alert' : 'status'} style={mensaje.tipo === 'ok' ? { marginBottom: 16 } : undefined}>{mensaje.texto}</div>}
            <div className="campo"><label htmlFor="e-nombre">Nombre completo</label><input id="e-nombre" type="text" value={nombre} onChange={(e) => setNombre(e.target.value)} /></div>
            <div className="campo"><label htmlFor="e-correo">Correo</label><input id="e-correo" type="email" value={correo} onChange={(e) => setCorreo(e.target.value)} /></div>
            <div className="campo">
              <label htmlFor="e-area">Área</label>
              <input id="e-area" type="text" list="lista-areas" value={area} onChange={(e) => setArea(e.target.value)} />
              <datalist id="lista-areas">{areasExistentes.map((a) => <option key={a} value={a} />)}</datalist>
              <span className="ayuda">Elige una existente o escribe una nueva.</span>
            </div>
            <button className="boton ancho" type="submit" disabled={guardando || !nombre || !correo || !area}>{guardando ? 'Guardando…' : 'Agregar empleado'}</button>
          </form>
        </aside>
      </div>
    </>
  );
}
