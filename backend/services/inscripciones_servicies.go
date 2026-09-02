// Archivo: services/inscripciones_services.go
package services

import (
	"backend/clients"
	"backend/dao"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CrearInscripcion valida y crea una nueva inscripción, y devuelve
// la entidad completa (incluyendo su ID) o un error.
func CrearInscripcion(usuarioID, horarioID uint) (dao.Inscripcion, error) {
	// 1) Verifica que el horario exista y traiga su actividad
	var horario dao.Horario
	if err := clients.DB.
		Preload("Actividad").
		First(&horario, horarioID).
		Error; err != nil {
		return dao.Inscripcion{}, fmt.Errorf("horario no encontrado")
	}

	// 2) Comprueba que la actividad esté activa
	if !horario.Actividad.Activo {
		return dao.Inscripcion{}, errors.New("la actividad está inactiva")
	}

	// 3) Cuenta cuántos usuarios hay ya inscritos en este horario
	var conteo int64
	if err := clients.DB.
		Model(&dao.Inscripcion{}).
		Where("horario_id = ?", horarioID).
		Count(&conteo).
		Error; err != nil {
		return dao.Inscripcion{}, fmt.Errorf("error contando inscripciones: %w", err)
	}

	// 4) Verifica si ya alcanzó el cupo máximo
	if uint16(conteo) >= horario.Actividad.CupoMax {
		return dao.Inscripcion{}, errors.New("no hay cupos disponibles")
	}

	// 5) Verifica si el usuario ya está inscrito
	var existente dao.Inscripcion
	err := clients.DB.
		Where("usuario_id = ? AND horario_id = ?", usuarioID, horarioID).
		First(&existente).
		Error
	// Si hubo un error distinto a "record not found", lo consideramos falla
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dao.Inscripcion{}, fmt.Errorf("error comprobando inscripción: %w", err)
	}
	// Si err == nil → ya existe inscripción
	if err == nil {
		return dao.Inscripcion{}, errors.New("ya estás inscrito a este horario")
	}

	// 6) Crea la inscripción
	insc := dao.Inscripcion{
		UsuarioID: usuarioID,
		HorarioID: horarioID,
	}
	if err := clients.DB.Create(&insc).Error; err != nil {
		return dao.Inscripcion{}, fmt.Errorf("error creando la inscripción: %w", err)
	}

	// 7) Devuelve la inscripción con su ID y HorarioID
	return insc, nil
}

// ListarInscripcionesPorUsuario devuelve las inscripciones (con horario y actividad) a las que un usuario está inscrito.
func ListarInscripcionesPorUsuario(usuarioID uint) ([]dao.Inscripcion, error) {
	var inscripciones []dao.Inscripcion

	err := clients.DB.
		Preload("Horario.Actividad").
		Where("usuario_id = ?", usuarioID).
		Find(&inscripciones).Error
	if err != nil {
		return nil, fmt.Errorf("error consultando inscripciones del usuario: %w", err)
	}

	return inscripciones, nil
}

// EliminarInscripcion borra una inscripción existente por su ID.
func EliminarInscripcion(inscID uint) error {
	if err := clients.DB.Delete(&dao.Inscripcion{}, inscID).Error; err != nil {
		return fmt.Errorf("error eliminando inscripción: %w", err)
	}
	return nil
}
