import { ArrowLeft } from 'lucide-react';
import { FormEvent, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ErrorApi, llamar } from '../api/cliente';
import { AreaConteo, CampanaBase, Plantilla } from '../api/tipos';
import { Esqueleto } from '../componentes/Estados';
import { useConsulta } from '../hooks/useConsulta';
import { numero } from '../utilidades/formato';

export function NuevaCampana() {
  const navegar = useNavigate();
  const plantillas = useConsulta(() => llamar<Plantilla[]>('/api/plantillas'), []);
  const areas = useConsulta(() => llamar<AreaConteo[]>('/api/areas'), []);

  const [nombre, setNombre] = useState('');
  const [plantillaId, setPlantillaId] = useState('');
  const [seleccion, setSeleccion] = useState<string[]>([]); // vacío = todas las áreas
  const [error, setError] = useState('');
  const [enviando, setEnviando] = useState(false);

  const totalDestinatarios = useMemo(() => {
    if (!areas.datos) return 0;
    return areas.datos.filter((a) => seleccion.length === 0 || seleccion.includes(a.area)).reduce((s, a) => s + a.empleados, 0);
  }, [areas.datos, seleccion]);

  const alternarArea = (area: string) =>
    setSeleccion((previo) => (previo.includes(area) ? previo.filter((a) => a !== area) : [...previo, area]));

  async function alEnviar(e: FormEvent) {
    e.preventDefault();
    setError('');
    if (nombre.trim().length < 3) return setError('Escribe un nombre de al menos 3 caracteres.');
    if (!plantillaId) return setError('Elige el tipo de ataque que quieres simular.');
    setEnviando(true);
    try {
      const campana = await llamar<CampanaBase>('/api/campanas', { metodo: 'POST', cuerpo: { nombre: nombre.trim(), plantilla_id: plantillaId, areas: seleccion } });
      navegar(`/campanas/${campana.id}`);
    } catch (err) {
      setError(err instanceof ErrorApi ? err.message : 'No se pudo crear la campaña.');
      setEnviando(false);
    }
  }

  const plantillaElegida = plantillas.datos?.find((p) => p.id === plantillaId);

  return (
    <>
      <header className="cabecera">
        <div>
          <Link to="/campanas" className="migas"><ArrowLeft size={14} aria-hidden="true" /> Campañas</Link>
          <h1>Nueva campaña</h1>
          <p>Los correos se envían al buzón de pruebas, no a direcciones reales.</p>
        </div>
      </header>

      <form onSubmit={alEnviar} noValidate>
        <div className="rejilla principal-lateral">
          <div className="tarjeta">
            {error && <div className="error-formulario" role="alert">{error}</div>}

            <div className="campo">
              <label htmlFor="nombre">Nombre de la campaña</label>
              <input id="nombre" type="text" maxLength={120} placeholder="Ej.: Simulacro de contraseñas · noviembre" value={nombre} onChange={(e) => setNombre(e.target.value)} />
            </div>

            <fieldset className="campo">
              <legend>Tipo de ataque</legend>
              {plantillas.cargando ? <Esqueleto alto={200} /> : (
                <div className="opciones-plantilla">
                  {plantillas.datos?.map((p) => (
                    <label className="opcion-plantilla" key={p.id}>
                      <input type="radio" name="plantilla" value={p.id} checked={plantillaId === p.id} onChange={() => setPlantillaId(p.id)} />
                      <span className="contenido">
                        <span className="linea"><strong>{p.nombre}</strong><span className="etiqueta tipo">{p.tipo}</span></span>
                        <span className="suave">{p.descripcion}</span>
                      </span>
                    </label>
                  ))}
                </div>
              )}
            </fieldset>

            <fieldset className="campo">
              <legend>Áreas que recibirán el correo</legend>
              <span className="ayuda">Si no eliges ninguna, se envía a todos los empleados.</span>
              {areas.cargando ? <Esqueleto alto={40} /> : (
                <div className="fichas" style={{ marginTop: 6 }}>
                  {areas.datos?.map((a) => (
                    <label className="ficha" key={a.area}>
                      <input type="checkbox" checked={seleccion.includes(a.area)} onChange={() => alternarArea(a.area)} />
                      <span>{a.area} · {a.empleados}</span>
                    </label>
                  ))}
                </div>
              )}
            </fieldset>
          </div>

          <aside className="tarjeta" aria-label="Resumen de la campaña">
            <h2>Resumen</h2>
            <dl className="resumen-lista">
              <div><dt>Destinatarios</dt><dd>{numero(totalDestinatarios)}</dd></div>
              <div><dt>Áreas</dt><dd>{seleccion.length === 0 ? 'Todas' : seleccion.length}</dd></div>
              <div><dt>Tipo de ataque</dt><dd>{plantillaElegida?.nombre ?? 'Sin elegir'}</dd></div>
              {plantillaElegida && <div><dt>Asunto del correo</dt><dd>{plantillaElegida.asunto}</dd></div>}
            </dl>
            <button className="boton ancho" type="submit" disabled={enviando || totalDestinatarios === 0}>
              {enviando ? 'Lanzando…' : 'Lanzar campaña'}
            </button>
            <p className="suave" style={{ marginTop: 12 }}>Al lanzar, los correos salen en segundos y verás los resultados en vivo.</p>
          </aside>
        </div>
      </form>
    </>
  );
}
