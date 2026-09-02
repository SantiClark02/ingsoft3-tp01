// dao/inscripciones.go
package dao

import "time"

// Inscripcion representa la inscripción de un usuario a un horario específico.
type Inscripcion struct {
	ID               uint      `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	UsuarioID        uint      `gorm:"column:usuario_id;not null"            json:"usuario_id"`
	HorarioID        uint      `gorm:"column:horario_id;not null"            json:"horario_id"`
	FechaInscripcion time.Time `gorm:"column:fecha_inscripcion;autoCreateTime" json:"fecha_inscripcion"`

	Usuario Usuario `gorm:"foreignKey:UsuarioID;references:ID" json:"usuario"`
	Horario Horario `gorm:"foreignKey:HorarioID;references:ID" json:"horario"`
}

func (Inscripcion) TableName() string {
	return "inscripciones"
}
