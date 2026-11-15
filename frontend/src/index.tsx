import '@fontsource-variable/bricolage-grotesque';
import '@fontsource-variable/figtree';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App } from './App';
import { ProveedorAuth } from './autenticacion/ContextoAuth';
import './estilos.css';

createRoot(document.getElementById('raiz') as HTMLElement).render(
  <BrowserRouter>
    <ProveedorAuth>
      <App />
    </ProveedorAuth>
  </BrowserRouter>,
);
