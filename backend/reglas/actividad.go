package reglas

import (
	"errors"
	"strings"
)

var (
	ErrTituloVacio  = errors.New("el título es obligatorio")
	ErrCupoInvalido = errors.New("el cupo máximo tiene que ser mayor a cero")
)

// ValidarActividad aplica las reglas mínimas para crear o editar una actividad.
// Función pura: se testea sin base de datos.
func ValidarActividad(titulo string, cupoMax int) error {
	if strings.TrimSpace(titulo) == "" {
		return ErrTituloVacio
	}
	if cupoMax <= 0 {
		return ErrCupoInvalido
	}
	return nil
}
