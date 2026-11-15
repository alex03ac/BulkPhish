import { createContext, ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { borrarSesion, EVENTO_SESION_EXPIRADA, guardarSesion, leerSesion, llamar, Sesion } from '../api/cliente';

interface ValorAuth {
  sesion: Sesion | null;
  iniciarSesion: (correo: string, clave: string) => Promise<void>;
  cerrarSesion: () => void;
}

const Contexto = createContext<ValorAuth | null>(null);

export function ProveedorAuth({ children }: { children: ReactNode }) {
  const [sesion, setSesion] = useState<Sesion | null>(() => leerSesion());

  const iniciarSesion = useCallback(async (correo: string, clave: string) => {
    const nueva = await llamar<Sesion>('/api/auth/login', { metodo: 'POST', cuerpo: { correo, clave }, publica: true });
    guardarSesion(nueva);
    setSesion(nueva);
  }, []);

  const cerrarSesion = useCallback(() => {
    borrarSesion();
    setSesion(null);
  }, []);

  // Si la API responde 401, el cliente avisa con un evento y aquí se cierra la sesión.
  useEffect(() => {
    const alExpirar = () => setSesion(null);
    window.addEventListener(EVENTO_SESION_EXPIRADA, alExpirar);
    return () => window.removeEventListener(EVENTO_SESION_EXPIRADA, alExpirar);
  }, []);

  const valor = useMemo(() => ({ sesion, iniciarSesion, cerrarSesion }), [sesion, iniciarSesion, cerrarSesion]);
  return <Contexto.Provider value={valor}>{children}</Contexto.Provider>;
}

export function useAuth(): ValorAuth {
  const valor = useContext(Contexto);
  if (!valor) throw new Error('useAuth debe usarse dentro de ProveedorAuth');
  return valor;
}
