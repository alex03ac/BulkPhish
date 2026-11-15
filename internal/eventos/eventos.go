// Paquete eventos: contrato de los mensajes que viajan por RabbitMQ.
// Es el único acuerdo entre servicios: ninguno conoce el código de otro.
package eventos

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

// Nombres de los intercambios (exchanges) de RabbitMQ.
const (
	Intercambio        = "bulkphish.eventos" // tipo topic: enruta por clave
	IntercambioMuertos = "bulkphish.muertos" // tipo fanout: recibe mensajes que fallaron
	ColaMuertos        = "bulkphish.muertos.cola"
)

// Claves de enrutamiento (routing keys): una por tipo de evento.
const (
	CampanaCreada   = "campana.creada"
	CorreoEnviado   = "correo.enviado"
	EnlaceClicado   = "enlace.clicado"
	CorreoReportado = "correo.reportado"
)

// Sobre es el formato común de todos los eventos.
type Sobre struct {
	ID         string          `json:"id"`
	Tipo       string          `json:"tipo"`
	OcurridoEn time.Time       `json:"ocurrido_en"`
	Datos      json.RawMessage `json:"datos"`
}

// NuevoSobre empaqueta unos datos dentro de un sobre con id único.
func NuevoSobre(tipo string, datos any) (Sobre, error) {
	crudo, err := json.Marshal(datos)
	if err != nil {
		return Sobre{}, err
	}
	return Sobre{ID: NuevoID(), Tipo: tipo, OcurridoEn: time.Now().UTC(), Datos: crudo}, nil
}

// Decodificar lee los datos del sobre dentro de la estructura destino.
func (s Sobre) Decodificar(destino any) error {
	return json.Unmarshal(s.Datos, destino)
}

// DestinatarioCampana es una persona que recibirá el correo simulado.
type DestinatarioCampana struct {
	DestinatarioID string `json:"destinatario_id"`
	Token          string `json:"token"`
	Nombre         string `json:"nombre"`
	Correo         string `json:"correo"`
	Area           string `json:"area"`
}

// DatosCampanaCreada viaja con la clave "campana.creada".
// Lleva todo lo necesario para que los consumidores no tengan que consultar
// la base de datos de campañas (transferencia de estado por evento).
type DatosCampanaCreada struct {
	CampanaID     string                `json:"campana_id"`
	Nombre        string                `json:"nombre"`
	PlantillaID   string                `json:"plantilla_id"`
	CreadaEn      time.Time             `json:"creada_en"`
	Destinatarios []DestinatarioCampana `json:"destinatarios"`
}

// DatosCorreoEnviado viaja con la clave "correo.enviado".
type DatosCorreoEnviado struct {
	CampanaID      string    `json:"campana_id"`
	DestinatarioID string    `json:"destinatario_id"`
	EnviadoEn      time.Time `json:"enviado_en"`
}

// DatosInteraccion sirve para "enlace.clicado" y "correo.reportado".
type DatosInteraccion struct {
	CampanaID      string    `json:"campana_id"`
	DestinatarioID string    `json:"destinatario_id"`
	OcurridoEn     time.Time `json:"ocurrido_en"`
}

// NuevoID genera un UUID versión 4.
func NuevoID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// NuevoSobreCampana construye el evento "campana.creada" con la lista de destinatarios.
func NuevoSobreCampana(campanaID, nombre, plantillaID string, creadaEn time.Time, destinatarios []DestinatarioCampana) (Sobre, error) {
	return NuevoSobre(CampanaCreada, DatosCampanaCreada{
		CampanaID: campanaID, Nombre: nombre, PlantillaID: plantillaID,
		CreadaEn: creadaEn.UTC(), Destinatarios: destinatarios,
	})
}
