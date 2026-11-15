package plantillas

import (
	"strings"
	"testing"
)

func TestTodasLasPlantillasRenderizan(t *testing.T) {
	for _, p := range Todas() {
		html, err := p.Renderizar(DatosCorreo{Nombre: "Lucía", EnlaceTrampa: "http://x/t/abc", EnlaceReporte: "http://x/r/abc"})
		if err != nil {
			t.Fatalf("%s: %v", p.ID, err)
		}
		if !strings.Contains(html, "http://x/t/abc") || !strings.Contains(html, "http://x/r/abc") {
			t.Errorf("%s: faltan los enlaces de clic o de reporte", p.ID)
		}
		if len(p.Senales) < 3 || p.Consejo == "" {
			t.Errorf("%s: falta contenido de capacitación", p.ID)
		}
	}
}

func TestEscapaDatosDelDestinatario(t *testing.T) {
	p, _ := Buscar("restablecer-clave")
	html, _ := p.Renderizar(DatosCorreo{Nombre: "<script>alert(1)</script>"})
	if strings.Contains(html, "<script>") {
		t.Fatal("el nombre no se escapó: riesgo de inyección de HTML")
	}
}
