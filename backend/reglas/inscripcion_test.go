package reglas

import (
	"errors"
	"testing"
)

// ---------- Regla 1: validación de la inscripción (función pura) ----------

// Parametrizado (table-driven): mismo comportamiento, varios datos.
func TestValidarInscripcion_CasosQueSeRechazan(t *testing.T) {
	casos := []struct {
		nombre      string
		activa      bool
		inscriptos  int64
		cupoMax     uint16
		yaInscripto bool
		esperado    error
	}{
		{"actividad inactiva", false, 0, 10, false, ErrActividadInactiva},
		{"cupo justo lleno (borde)", true, 10, 10, false, ErrSinCupo},
		{"cupo excedido", true, 12, 10, false, ErrSinCupo},
		{"usuario ya inscripto", true, 3, 10, true, ErrYaInscripto},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := ValidarInscripcion(c.activa, c.inscriptos, c.cupoMax, c.yaInscripto)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("esperaba %v, obtuve %v", c.esperado, err)
			}
		})
	}
}

// Caso de borde: queda exactamente un lugar.
func TestValidarInscripcion_UltimoCupoDisponible_EsAceptada(t *testing.T) {
	err := ValidarInscripcion(true, 9, 10, false)

	if err != nil {
		t.Fatalf("con 9 de 10 inscriptos debería aceptar, obtuve %v", err)
	}
}

// ---------- Regla 2: el servicio de inscripción (con mock) ----------

// repoMock reemplaza a la base: contesta datos fijos (stub)
// y registra cuántas veces se llamó a Crear (eso lo vuelve un mock).
type repoMock struct {
	activa        bool
	cupoMax       uint16
	inscriptos    int64
	yaInscripto   bool
	errHorario    error
	llamadasCrear int
}

func (m *repoMock) DatosHorario(uint) (bool, uint16, error) {
	return m.activa, m.cupoMax, m.errHorario
}
func (m *repoMock) ContarInscriptos(uint) (int64, error)  { return m.inscriptos, nil }
func (m *repoMock) YaInscripto(uint, uint) (bool, error) { return m.yaInscripto, nil }
func (m *repoMock) Crear(uint, uint) (uint, error) {
	m.llamadasCrear++
	return 99, nil
}

func TestInscribir_TodoValido_CreaUnaSolaVez(t *testing.T) {
	// Arrange
	repo := &repoMock{activa: true, cupoMax: 10, inscriptos: 2}

	// Act
	id, err := Inscribir(repo, 1, 5)

	// Assert: no sólo el valor devuelto, también la INTERACCIÓN con el doble
	if err != nil || id != 99 {
		t.Fatalf("esperaba id 99 sin error, obtuve %d, %v", id, err)
	}
	if repo.llamadasCrear != 1 {
		t.Fatalf("Crear debía llamarse 1 vez, se llamó %d", repo.llamadasCrear)
	}
}

func TestInscribir_SinCupo_NoCreaLaInscripcion(t *testing.T) {
	repo := &repoMock{activa: true, cupoMax: 10, inscriptos: 10}

	_, err := Inscribir(repo, 1, 5)

	if !errors.Is(err, ErrSinCupo) {
		t.Fatalf("esperaba ErrSinCupo, obtuve %v", err)
	}
	if repo.llamadasCrear != 0 {
		t.Fatalf("con el cupo lleno no debía crear nada, se llamó %d veces", repo.llamadasCrear)
	}
}

// Caso de error: el horario no existe.
func TestInscribir_HorarioInexistente_DevuelveError(t *testing.T) {
	repo := &repoMock{errHorario: errors.New("record not found")}

	_, err := Inscribir(repo, 1, 404)

	if !errors.Is(err, ErrHorarioInexistente) {
		t.Fatalf("esperaba ErrHorarioInexistente, obtuve %v", err)
	}
}
