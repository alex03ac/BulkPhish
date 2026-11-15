import { ShieldAlert, ShieldCheck, X } from 'lucide-react';
import { useParams } from 'react-router-dom';
import { llamar } from '../api/cliente';
import { Capacitacion as DatosCapacitacion } from '../api/tipos';
import { Esqueleto } from '../componentes/Estados';
import { Isotipo } from '../componentes/Marca';
import { useConsulta } from '../hooks/useConsulta';

function Marco({ children }: { children: React.ReactNode }) {
  return (
    <div className="publica">
      <div className="publica-marca"><Isotipo tamano={28} /> BulkPhish</div>
      <main className="publica-tarjeta">{children}</main>
    </div>
  );
}

/**
 * Página pública (sin sesión) a la que llega la persona al hacer clic ("capacitacion")
 * o al reportar el correo ("reportado"). El token del enlace identifica a la persona.
 */
export function Capacitacion({ modo }: { modo: 'clic' | 'reporte' }) {
  const { token = '' } = useParams();
  const { datos, error, cargando } = useConsulta(() => llamar<DatosCapacitacion>(`/api/publico/capacitacion/${token}`, { publica: true }), [token]);

  if (cargando) return <Marco><Esqueleto alto={220} /></Marco>;
  if (error || !datos) {
    return (
      <Marco>
        <h1>Este enlace no es válido</h1>
        <p>Puede haber expirado o estar incompleto. Si recibiste un correo y tienes dudas, consulta con el equipo de seguridad de tu empresa.</p>
      </Marco>
    );
  }

  if (modo === 'reporte') {
    return (
      <Marco>
        <span className="sello bien"><ShieldCheck size={16} aria-hidden="true" /> Reporte recibido</span>
        <h1>Gracias por reportarlo, {datos.nombre}</h1>
        <p>Ese correo era un simulacro de phishing de tu empresa, y lo detectaste a tiempo. Reportar los mensajes sospechosos es la mejor defensa: permite que el equipo de seguridad actúe antes de que alguien más caiga.</p>
        <h2>Las señales que lo delataban</h2>
        <ul className="senales">
          {datos.plantilla.senales.map((s) => (<li key={s}><X size={18} aria-hidden="true" /><span>{s}</span></li>))}
        </ul>
        <div className="consejo">{datos.plantilla.consejo}</div>
      </Marco>
    );
  }

  return (
    <Marco>
      <span className="sello aviso"><ShieldAlert size={16} aria-hidden="true" /> Esto era un simulacro</span>
      <h1>No pasó nada, {datos.nombre}. Pero sirve para aprender.</h1>
      <p>El correo que abriste forma parte de un ejercicio de seguridad de tu empresa ({datos.plantilla.tipo.toLowerCase()}). No se instaló nada en tu equipo ni se guardó ninguna contraseña. Un atacante real sí lo habría intentado, y por eso practicamos.</p>
      <h2>Qué podías notar</h2>
      <ul className="senales">
        {datos.plantilla.senales.map((s) => (<li key={s}><X size={18} aria-hidden="true" /><span>{s}</span></li>))}
      </ul>
      <div className="consejo"><strong>La próxima vez:</strong> {datos.plantilla.consejo}</div>
      <p className="suave" style={{ marginTop: 18 }}>Si dudas de un mensaje, usa el botón <strong>Reportar phishing</strong> de tu correo. Ya puedes cerrar esta página.</p>
    </Marco>
  );
}
