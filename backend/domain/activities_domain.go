package domain

// Horario representa un turno específico de una actividad.
type Horarios struct {
	ID          uint   `json:"id"`
	ActividadID uint   `json:"actividad_id"`
	Day         string `json:"day"`        // Ej: "Lun", "Mar", etc.
	StartTime   string `json:"start_time"` // Ej: "18:00"
	DurationMin uint16 `json:"duration"`   // Duración en minutos
}

// Activity representa una actividad con todos sus horarios.
type Activity struct {
	ID                  uint    `json:"id"`
	Name                string  `json:"name"`        // Ej: "Zumba", "Musculación"
	Description         string  `json:"description"` // Detalle de la actividad
	Capacity            uint16  `json:"capacity"`    // Cupo máximo
	ImageURL            *string `json:"image_url,omitempty"`
	CategoryName        string  `json:"category_name"`
	InstructorName      string  `json:"instructor_name"`
	InstructorEmail     *string `json:"instructor_email,omitempty"`
	InstructorSpecialty *string `json:"instructor_specialty,omitempty"`
	Active              bool    `json:"active"` // Indica si la actividad está habilitada

	// Lista de turnos disponibles
	Schedule []Horario `json:"schedule"`
}
