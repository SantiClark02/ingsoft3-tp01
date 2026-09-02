package dao

type Actividad struct {
	ID          uint    `gorm:"primaryKey;column:id;autoIncrement"`
	Titulo      string  `gorm:"column:titulo;not null"`
	Descripcion string  `gorm:"column:descripcion;type:text;not null"`
	CupoMax     uint16  `gorm:"column:cupo_max;not null"`
	ImagenURL   *string `gorm:"column:imagen_url;type:varchar(255)"`

	NombreCategoria        string  `gorm:"column:nombre_categoria;type:varchar(50);not null"`
	NombreInstructor       string  `gorm:"column:nombre_instructor;type:varchar(100);not null"`
	EmailInstructor        *string `gorm:"column:email_instructor;type:varchar(150)"`
	EspecialidadInstructor *string `gorm:"column:especialidad_instructor;type:varchar(100)"`

	Activo bool `gorm:"column:activo;default:true"`

	// Una actividad puede tener múltiples horarios
	Horarios []Horario `gorm:"foreignKey:ActividadID"`
}

func (Actividad) TableName() string {
	return "actividades"
}
