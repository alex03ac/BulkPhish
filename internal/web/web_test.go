package web

import "testing"

func TestEsUUID(t *testing.T) {
	if !EsUUID("123e4567-e89b-42d3-a456-426614174000") {
		t.Error("UUID válido rechazado")
	}
	for _, malo := range []string{"", "no-es-uuid", "123e4567e89b42d3a456426614174000", "123e4567-e89b-42d3-a456-42661417400z"} {
		if EsUUID(malo) {
			t.Errorf("%q no debería ser UUID", malo)
		}
	}
}

func TestLosTokensNoQuedanEnLosRegistros(t *testing.T) {
	casos := map[string]string{
		"/t/0123456789abcdef0123456789abcdef":                        "/t/{id}",
		"/api/publico/capacitacion/0123456789abcdef0123456789abcdef": "/api/publico/capacitacion/{id}",
		"/api/campanas/123e4567-e89b-42d3-a456-426614174000":         "/api/campanas/{id}",
		"/api/resultados/resumen":                                    "/api/resultados/resumen",
	}
	for entrada, esperado := range casos {
		if obtenido := rutaSegura(entrada); obtenido != esperado {
			t.Errorf("rutaSegura(%q) = %q, esperado %q", entrada, obtenido, esperado)
		}
	}
}
