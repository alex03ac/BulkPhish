package autenticacion

import (
	"sync"
	"time"
)

// limitador frena los ataques de fuerza bruta: tras demasiados intentos fallidos
// seguidos desde la misma IP para el mismo correo, bloquea nuevos intentos un tiempo.
// Vive en memoria: es suficiente para una sola instancia. Con varias réplicas
// habría que guardar los contadores en un almacén compartido (por ejemplo Redis).
type limitador struct {
	mu      sync.Mutex
	fallos  map[string][]time.Time
	maximo  int
	ventana time.Duration
	ahora   func() time.Time
}

func nuevoLimitador(maximo int, ventana time.Duration) *limitador {
	return &limitador{fallos: map[string][]time.Time{}, maximo: maximo, ventana: ventana, ahora: time.Now}
}

// recientes devuelve los fallos que siguen dentro de la ventana (y descarta los viejos).
func (l *limitador) recientes(clave string) []time.Time {
	limite := l.ahora().Add(-l.ventana)
	vigentes := l.fallos[clave][:0]
	for _, t := range l.fallos[clave] {
		if t.After(limite) {
			vigentes = append(vigentes, t)
		}
	}
	if len(vigentes) == 0 {
		delete(l.fallos, clave)
	} else {
		l.fallos[clave] = vigentes
	}
	return vigentes
}

// bloqueado indica si la clave ya agotó sus intentos.
func (l *limitador) bloqueado(clave string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.recientes(clave)) >= l.maximo
}

func (l *limitador) registrarFallo(clave string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.recientes(clave)
	l.fallos[clave] = append(l.fallos[clave], l.ahora())
}

func (l *limitador) reiniciar(clave string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fallos, clave)
}
