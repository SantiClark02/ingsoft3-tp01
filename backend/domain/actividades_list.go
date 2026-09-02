package domain

// ActivityListItem representa la info para el listado,
// con un flag que indica si el usuario ya está inscrito,
// y, en ese caso, el ID de la inscripción para poder borrarla.
type ActivityListItem struct {
	ID             uint    `json:"id"`
	Title          string  `json:"title"`
	CategoryName   string  `json:"category_name"`
	InstructorName string  `json:"instructor_name"`
	Day            string  `json:"day"`
	StartTime      string  `json:"start_time"`
	Duration       uint16  `json:"duration"`
	Capacity       uint16  `json:"capacity"`
	ImageURL       *string `json:"image_url,omitempty"`

	// Estos dos son los que faltan:
	Subscribed    bool `json:"subscribed"`
	InscriptionID uint `json:"inscription_id,omitempty"`
}
