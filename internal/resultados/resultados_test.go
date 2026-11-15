package resultados

import "testing"

func TestCalcularTasas(t *testing.T) {
	e := calcular(10, 8, 2, 4)
	if e.TasaClic != 0.25 || e.TasaReporte != 0.5 {
		t.Fatalf("tasas incorrectas: %+v", e)
	}
}

func TestCalcularSinEnviadosNoDivideEntreCero(t *testing.T) {
	e := calcular(10, 0, 0, 0)
	if e.TasaClic != 0 || e.TasaReporte != 0 {
		t.Fatalf("debía ser 0: %+v", e)
	}
}
