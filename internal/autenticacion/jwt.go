package autenticacion

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const emisor = "bulkphish-auth"

// Reclamos es el contenido (payload) del token JWT.
type Reclamos struct {
	Nombre string `json:"nombre"`
	Correo string `json:"correo"`
	jwt.RegisteredClaims
}

// emisorJWT firma y verifica tokens con una clave secreta compartida (HS256).
type emisorJWT struct {
	secreto  []byte
	duracion time.Duration
}

// emitir crea un token firmado para el administrador indicado.
func (e *emisorJWT) emitir(id, nombre, correo string) (string, time.Time, error) {
	ahora := time.Now()
	expira := ahora.Add(e.duracion)
	reclamos := Reclamos{
		Nombre: nombre,
		Correo: correo,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id,
			Issuer:    emisor,
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(expira),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, reclamos).SignedString(e.secreto)
	return token, expira, err
}

// verificar valida firma, algoritmo, emisor y expiración.
func (e *emisorJWT) verificar(texto string) (*Reclamos, error) {
	reclamos := &Reclamos{}
	token, err := jwt.ParseWithClaims(texto, reclamos,
		func(*jwt.Token) (any, error) { return e.secreto, nil },
		jwt.WithValidMethods([]string{"HS256"}), // evita el ataque de "alg: none"
		jwt.WithIssuer(emisor),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token inválido")
	}
	return reclamos, nil
}
