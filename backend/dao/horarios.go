package dao

import "time"

type Horario struct {
	ID          uint      `gorm:"primaryKey;column:id;autoIncrement"`
	ActividadID uint      `gorm:"column:actividad_id;not null;index"`
	DiaSemana   string    `gorm:"column:dia_semana;type:enum('Lun','Mar','Mie','Jue','Vie','Sab','Dom');not null"`
	HoraInicio  time.Time `gorm:"column:hora_inicio;type:time;not null"`
	DuracionMin uint16    `gorm:"column:duracion_min;not null"`

	// Relación con Actividad
	Actividad Actividad `gorm:"foreignKey:ActividadID;references:ID"`
}

func (Horario) TableName() string {
	return "horarios"
}
