package reglas

import (
	"errors"
	"testing"
)

// ---------- Regla 5: datos mínimos de una actividad ----------

func TestValidarActividad_DatosInvalidos_SonRechazados(t *testing.T) {
	casos := []struct {
		nombre   string
		titulo   string
		cupo     int
		esperado error
	}{
		{"título vacío", "", 10, ErrTituloVacio},
		{"título sólo espacios", "   ", 10, ErrTituloVacio},
		{"cupo cero (borde)", "Yoga", 0, ErrCupoInvalido},
		{"cupo negativo", "Yoga", -5, ErrCupoInvalido},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := ValidarActividad(c.titulo, c.cupo)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("esperaba %v, obtuve %v", c.esperado, err)
			}
		})
	}
}

func TestValidarActividad_DatosValidos_EsAceptada(t *testing.T) {
	err := ValidarActividad("Yoga", 1)

	if err != nil {
		t.Fatalf("una actividad con título y cupo 1 debería aceptarse, obtuve %v", err)
	}
}
