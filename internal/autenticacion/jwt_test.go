package autenticacion

import (
	"testing"
	"time"
)

func TestEmitirYVerificar(t *testing.T) {
	e := &emisorJWT{secreto: []byte("secreto-de-prueba-con-mas-de-32-caracteres"), duracion: time.Hour}
	token, _, err := e.emitir("id-1", "Ana", "ana@ejemplo.test")
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.verificar(token)
	if err != nil || r.Correo != "ana@ejemplo.test" {
		t.Fatalf("token válido rechazado: %v", err)
	}
}

func TestRechazaTokenExpiradoOManipulado(t *testing.T) {
	e := &emisorJWT{secreto: []byte("secreto-de-prueba-con-mas-de-32-caracteres"), duracion: -time.Minute}
	vencido, _, _ := e.emitir("id-1", "Ana", "ana@ejemplo.test")
	if _, err := e.verificar(vencido); err == nil {
		t.Fatal("debió rechazar un token expirado")
	}
	otro := &emisorJWT{secreto: []byte("otro-secreto-distinto-de-mas-de-32-caracteres"), duracion: time.Hour}
	ajeno, _, _ := otro.emitir("id-2", "Eva", "eva@ejemplo.test")
	if _, err := e.verificar(ajeno); err == nil {
		t.Fatal("debió rechazar un token firmado con otra clave")
	}
}
