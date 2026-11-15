package campanas

import "testing"

func TestTokenValido(t *testing.T) {
	casos := map[string]bool{
		generarToken():                     true,
		"":                                 false,
		"abc":                              false,
		"ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ": false,
		"0123456789abcdef0123456789abcdef": true,
	}
	for entrada, esperado := range casos {
		if tokenValido(entrada) != esperado {
			t.Errorf("tokenValido(%q) debía ser %v", entrada, esperado)
		}
	}
}

func TestTokensSonUnicos(t *testing.T) {
	vistos := map[string]bool{}
	for i := 0; i < 1000; i++ {
		tk := generarToken()
		if vistos[tk] {
			t.Fatal("token repetido")
		}
		vistos[tk] = true
	}
}
