import { ReactNode } from 'react';
import { ErrorApi } from '../api/cliente';

export function Esqueleto({ alto = 120 }: { alto?: number }) {
  return <div className="esqueleto" style={{ height: alto }} aria-hidden="true" />;
}

export function Vacio({ titulo, children }: { titulo: string; children?: ReactNode }) {
  return (
    <div className="vacio">
      <h3>{titulo}</h3>
      {children}
    </div>
  );
}

export function ErrorCarga({ error, alReintentar }: { error: ErrorApi; alReintentar: () => void }) {
  return (
    <div className="tarjeta" role="alert">
      <Vacio titulo="No pudimos cargar esta información">
        <p>{error.message}</p>
        <button type="button" className="boton secundario" onClick={alReintentar}>Reintentar</button>
      </Vacio>
    </div>
  );
}
