package domain

// Horario representa un turno específico de una actividad.
type Horario struct {
	ID          uint   `json:"id"`
	ActividadID uint   `json:"actividad_id"`
	Day         string `json:"day"`        // "Lun", "Mar", "Mie", "Jue", "Vie", "Sab", "Dom"
	StartTime   string `json:"start_time"` // Hora, ej 18:30
	DurationMin uint16 `json:"duration"`   // Duración en minutos
}
