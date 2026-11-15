import { useCallback, useEffect, useRef, useState } from 'react';
import { ErrorApi } from '../api/cliente';

interface Estado<T> {
  datos: T | null;
  error: ErrorApi | null;
  cargando: boolean;
}

/**
 * Ejecuta una consulta al montar y, si se indica un intervalo, la repite
 * (solo mientras la pestaña está visible). Conserva los datos anteriores si una repetición falla.
 */
export function useConsulta<T>(consulta: () => Promise<T>, dependencias: unknown[], intervaloMs = 0) {
  const [estado, setEstado] = useState<Estado<T>>({ datos: null, error: null, cargando: true });
  const consultaRef = useRef(consulta);
  consultaRef.current = consulta;

  const ejecutar = useCallback(async () => {
    try {
      const datos = await consultaRef.current();
      setEstado({ datos, error: null, cargando: false });
    } catch (e) {
      const error = e instanceof ErrorApi ? e : new ErrorApi(0, 'Ocurrió un error inesperado.');
      setEstado((previo) => ({ datos: previo.datos, error, cargando: false }));
    }
  }, []);

  useEffect(() => {
    setEstado({ datos: null, error: null, cargando: true });
    void ejecutar();
    if (!intervaloMs) return;
    const reloj = setInterval(() => {
      if (document.visibilityState === 'visible') void ejecutar();
    }, intervaloMs);
    return () => clearInterval(reloj);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...dependencias, intervaloMs]);

  return { ...estado, recargar: ejecutar };
}
