package autenticacion

import (
	"testing"
	"time"
)

func TestBloqueaTrasVariosFallosYSeLiberaConElTiempo(t *testing.T) {
	ahora := time.Now()
	l := nuevoLimitador(3, 10*time.Minute)
	l.ahora = func() time.Time { return ahora }

	for i := 0; i < 3; i++ {
		if l.bloqueado("a|1.1.1.1") {
			t.Fatalf("no debía bloquear antes del intento %d", i+1)
		}
		l.registrarFallo("a|1.1.1.1")
	}
	if !l.bloqueado("a|1.1.1.1") {
		t.Fatal("debía bloquear tras 3 fallos")
	}
	if l.bloqueado("b|1.1.1.1") {
		t.Fatal("otro correo no debe verse afectado")
	}
	ahora = ahora.Add(11 * time.Minute)
	if l.bloqueado("a|1.1.1.1") {
		t.Fatal("debía liberarse al pasar la ventana")
	}
}

func TestUnLoginCorrectoReiniciaElContador(t *testing.T) {
	l := nuevoLimitador(2, time.Minute)
	l.registrarFallo("x")
	l.reiniciar("x")
	l.registrarFallo("x")
	if l.bloqueado("x") {
		t.Fatal("el contador debía haberse reiniciado")
	}
}
