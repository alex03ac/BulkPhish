import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { useAuth } from './autenticacion/ContextoAuth';
import { Estructura } from './componentes/Estructura';
import { Campanas } from './paginas/Campanas';
import { Capacitacion } from './paginas/Capacitacion';
import { DetalleCampana } from './paginas/DetalleCampana';
import { Empleados } from './paginas/Empleados';
import { Login } from './paginas/Login';
import { NoEncontrada } from './paginas/NoEncontrada';
import { NuevaCampana } from './paginas/NuevaCampana';
import { Panel } from './paginas/Panel';

/** Solo deja pasar a quien tiene sesión; si no, lo envía al inicio de sesión y recuerda a dónde iba. */
function RutaProtegida() {
  const { sesion } = useAuth();
  const ubicacion = useLocation();
  if (!sesion) return <Navigate to="/ingresar" replace state={{ desde: ubicacion.pathname }} />;
  return <Estructura />;
}

export function App() {
  return (
    <Routes>
      {/* Públicas */}
      <Route path="/ingresar" element={<Login />} />
      <Route path="/capacitacion/:token" element={<Capacitacion modo="clic" />} />
      <Route path="/reportado/:token" element={<Capacitacion modo="reporte" />} />

      {/* Protegidas (solo administradores) */}
      <Route element={<RutaProtegida />}>
        <Route path="/" element={<Panel />} />
        <Route path="/campanas" element={<Campanas />} />
        <Route path="/campanas/nueva" element={<NuevaCampana />} />
        <Route path="/campanas/:id" element={<DetalleCampana />} />
        <Route path="/empleados" element={<Empleados />} />
      </Route>
      <Route path="*" element={<NoEncontrada />} />
    </Routes>
  );
}
