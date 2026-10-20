// Archivo: services/inscripciones_services.go
package services

import (
	"backend/clients"
	"backend/dao"
	"backend/reglas"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// repoGorm es la implementación REAL del repositorio que pide reglas.Inscribir:
// habla con MySQL a través de GORM. En los tests se usa un doble en su lugar.
type repoGorm struct{}

func (repoGorm) DatosHorario(horarioID uint) (bool, uint16, error) {
	var horario dao.Horario
	if err := clients.DB.Preload("Actividad").First(&horario, horarioID).Error; err != nil {
		return false, 0, err
	}
	return horario.Actividad.Activo, horario.Actividad.CupoMax, nil
}

func (repoGorm) ContarInscriptos(horarioID uint) (int64, error) {
	var conteo int64
	if err := clients.DB.Model(&dao.Inscripcion{}).Where("horario_id = ?", horarioID).Count(&conteo).Error; err != nil {
		return 0, fmt.Errorf("error contando inscripciones: %w", err)
	}
	return conteo, nil
}

func (repoGorm) YaInscripto(usuarioID, horarioID uint) (bool, error) {
	var existente dao.Inscripcion
	err := clients.DB.Where("usuario_id = ? AND horario_id = ?", usuarioID, horarioID).First(&existente).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error comprobando inscripción: %w", err)
	}
	return true, nil
}

func (repoGorm) Crear(usuarioID, horarioID uint) (uint, error) {
	insc := dao.Inscripcion{UsuarioID: usuarioID, HorarioID: horarioID}
	if err := clients.DB.Create(&insc).Error; err != nil {
		return 0, fmt.Errorf("error creando la inscripción: %w", err)
	}
	return insc.ID, nil
}

// CrearInscripcion conserva la firma que usa el controller: delega las reglas
// en reglas.Inscribir y le pasa el repositorio real (acá se "cablea" la dependencia).
func CrearInscripcion(usuarioID, horarioID uint) (dao.Inscripcion, error) {
	id, err := reglas.Inscribir(repoGorm{}, usuarioID, horarioID)
	if err != nil {
		return dao.Inscripcion{}, err
	}
	return dao.Inscripcion{ID: id, UsuarioID: usuarioID, HorarioID: horarioID}, nil
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
