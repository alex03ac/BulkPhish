import { FormEvent, useState } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../autenticacion/ContextoAuth';
import { ErrorApi } from '../api/cliente';
import { Isotipo } from '../componentes/Marca';

// Mosaico decorativo: muestra la idea del producto (quién hizo qué) sin usar números.
const PATRON = 'eeceeeeeeeeeeeeeereeeeeeeeceeeeeeeeeeceeeeeeeeeeeeeeeeeeeeeereeeeeeeeeeeeeeeeceeeee'.padEnd(72, 'e');

export function Login() {
  const { sesion, iniciarSesion } = useAuth();
  const ubicacion = useLocation();
  const [correo, setCorreo] = useState('');
  const [clave, setClave] = useState('');
  const [error, setError] = useState('');
  const [enviando, setEnviando] = useState(false);

  if (sesion) return <Navigate to={(ubicacion.state as { desde?: string } | null)?.desde ?? '/'} replace />;

  async function alEnviar(e: FormEvent) {
    e.preventDefault();
    setError('');
    setEnviando(true);
    try {
      await iniciarSesion(correo, clave);
    } catch (err) {
      setError(err instanceof ErrorApi ? err.message : 'No se pudo iniciar sesión.');
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div className="acceso">
      <section className="acceso-marca">
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Isotipo tamano={34} />
          <span className="marca-nombre">BulkPhish</span>
        </div>
        <div>
          <h1>Descubre quién caería antes de que lo intente un atacante.</h1>
          <p>Lanza simulacros de phishing controlados, mide cómo responde cada equipo y capacita en el momento exacto del error.</p>
        </div>
        <div className="acceso-mosaico" aria-hidden="true">
          {PATRON.split('').map((c, i) => (
            <i key={i} className={c === 'c' ? 'c' : c === 'r' ? 'r' : i % 5 === 0 ? 'e' : ''} />
          ))}
        </div>
      </section>
      <section className="acceso-formulario">
        <form onSubmit={alEnviar} noValidate>
          <h2>Iniciar sesión</h2>
          <p className="suave">Acceso para administradores de seguridad.</p>
          {error && <div className="error-formulario" role="alert">{error}</div>}
          <div className="campo">
            <label htmlFor="correo">Correo</label>
            <input id="correo" type="email" autoComplete="username" required value={correo} onChange={(e) => setCorreo(e.target.value)} />
          </div>
          <div className="campo">
            <label htmlFor="clave">Contraseña</label>
            <input id="clave" type="password" autoComplete="current-password" required value={clave} onChange={(e) => setClave(e.target.value)} />
          </div>
          <button className="boton ancho" type="submit" disabled={enviando || !correo || !clave}>
            {enviando ? 'Verificando…' : 'Iniciar sesión'}
          </button>
        </form>
      </section>
    </div>
  );
}
