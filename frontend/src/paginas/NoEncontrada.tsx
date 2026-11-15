import { Link } from 'react-router-dom';

export function NoEncontrada() {
  return (
    <div className="publica">
      <main className="publica-tarjeta">
        <h1>Página no encontrada</h1>
        <p>La dirección que escribiste no existe o cambió de lugar.</p>
        <p style={{ marginTop: 18 }}><Link className="boton" to="/">Ir al panel</Link></p>
      </main>
    </div>
  );
}
