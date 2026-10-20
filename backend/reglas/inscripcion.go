// Package reglas concentra las reglas de negocio de la inscripción.
// No importa nada de la base de datos: por eso se puede testear sola.
package reglas

import "errors"

var (
	ErrHorarioInexistente = errors.New("horario no encontrado")
	ErrActividadInactiva  = errors.New("la actividad está inactiva")
	ErrSinCupo            = errors.New("no hay cupos disponibles")
	ErrYaInscripto        = errors.New("ya estás inscrito a este horario")
)

// ValidarInscripcion es una función pura: recibe datos y decide.
// No toca la base, no necesita dobles para testearse.
func ValidarInscripcion(activa bool, inscriptos int64, cupoMax uint16, yaInscripto bool) error {
	if !activa {
		return ErrActividadInactiva
	}
	if inscriptos >= int64(cupoMax) {
		return ErrSinCupo
	}
	if yaInscripto {
		return ErrYaInscripto
	}
	return nil
}

// Repositorio es lo que la inscripción necesita de la base.
// Entra desde afuera (inyección de dependencias): la app le pasa
// la implementación con GORM y los tests le pasan un doble.
type Repositorio interface {
	DatosHorario(horarioID uint) (activa bool, cupoMax uint16, err error)
	ContarInscriptos(horarioID uint) (int64, error)
	YaInscripto(usuarioID, horarioID uint) (bool, error)
	Crear(usuarioID, horarioID uint) (uint, error)
}

// Inscribir aplica las reglas y, sólo si se cumplen, crea la inscripción.
func Inscribir(repo Repositorio, usuarioID, horarioID uint) (uint, error) {
	activa, cupoMax, err := repo.DatosHorario(horarioID)
	if err != nil {
		return 0, ErrHorarioInexistente
	}
	inscriptos, err := repo.ContarInscriptos(horarioID)
	if err != nil {
		return 0, err
	}
	ya, err := repo.YaInscripto(usuarioID, horarioID)
	if err != nil {
		return 0, err
	}
	if err := ValidarInscripcion(activa, inscriptos, cupoMax, ya); err != nil {
		return 0, err
	}
	return repo.Crear(usuarioID, horarioID)
}
