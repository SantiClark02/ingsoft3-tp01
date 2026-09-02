package services

import (
	"backend/clients"
	"backend/dao"
	"backend/domain"
	"fmt"
)

// ListarHorariosPorActividad retorna todos los horarios asociados a una actividad dada.
func ListarHorariosPorActividad(actividadID uint) ([]domain.Horario, error) {
	var daoHorarios []dao.Horario
	if tx := clients.DB.Where("actividad_id = ?", actividadID).Find(&daoHorarios); tx.Error != nil {
		return nil, fmt.Errorf("error listando horarios: %w", tx.Error)
	}

	var horarios []domain.Horario
	for _, h := range daoHorarios {
		horarios = append(horarios, domain.Horario{
			ID:          h.ID,
			ActividadID: h.ActividadID,
			Day:         h.DiaSemana,
			StartTime:   h.HoraInicio.Format("15:04"),
			DurationMin: h.DuracionMin,
		})
	}
	return horarios, nil
}

// GetHorarioByID retorna un único horario por su ID.
func GetHorarioByID(id uint) (domain.Horario, error) {
	var h dao.Horario
	if tx := clients.DB.First(&h, id); tx.Error != nil {
		return domain.Horario{}, fmt.Errorf("horario con ID %d no encontrado", id)
	}
	return domain.Horario{
		ID:          h.ID,
		ActividadID: h.ActividadID,
		Day:         h.DiaSemana,
		StartTime:   h.HoraInicio.Format("15:04"),
		DurationMin: h.DuracionMin,
	}, nil
}
