package domain

import "time"

// Inscripcion representa la inscripción de un usuario a un horario (turno).
type Inscripcion struct {
	ID               uint      `json:"id"`
	UsuarioID        uint      `json:"usuario_id"`
	HorarioID        uint      `json:"horario_id"`
	FechaInscripcion time.Time `json:"fecha_inscripcion"`
}
