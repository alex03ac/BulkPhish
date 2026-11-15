import { LayoutDashboard, LogOut, Send, Users } from 'lucide-react';
import { NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../autenticacion/ContextoAuth';
import { Isotipo } from './Marca';

/** Estructura de las páginas del administrador: barra lateral + contenido. */
export function Estructura() {
  const { sesion, cerrarSesion } = useAuth();
  return (
    <div className="aplicacion">
      <aside className="lateral">
        <NavLink to="/" className="marca" aria-label="BulkPhish, ir al panel">
          <Isotipo />
          <span className="marca-nombre">BulkPhish</span>
        </NavLink>
        <nav className="nav" aria-label="Principal">
          <NavLink to="/" end className={({ isActive }) => (isActive ? 'activo' : '')}>
            <LayoutDashboard size={18} aria-hidden="true" /><span>Panel</span>
          </NavLink>
          <NavLink to="/campanas" className={({ isActive }) => (isActive ? 'activo' : '')}>
            <Send size={18} aria-hidden="true" /><span>Campañas</span>
          </NavLink>
          <NavLink to="/empleados" className={({ isActive }) => (isActive ? 'activo' : '')}>
            <Users size={18} aria-hidden="true" /><span>Empleados</span>
          </NavLink>
        </nav>
        <div className="lateral-pie">
          <strong>{sesion?.usuario.nombre}</strong>
          <span>{sesion?.usuario.correo}</span>
          <button type="button" className="salir" onClick={cerrarSesion}>
            <LogOut size={16} aria-hidden="true" /> Cerrar sesión
          </button>
        </div>
      </aside>
      <main className="principal" id="contenido">
        <Outlet />
      </main>
    </div>
  );
}
