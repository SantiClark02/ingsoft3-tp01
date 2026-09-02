package dao

import "time"

type Usuario struct {
	ID           uint      `gorm:"primaryKey;column:id;autoIncrement"`
	Username     string    `gorm:"column:nombre;type:varchar(100);not null"`
	Email        string    `gorm:"column:email;type:varchar(150);not null;unique"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(256);not null"`
	Rol          string    `gorm:"column:rol;type:enum('socio','administrador');not null"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`

	Inscripciones []Inscripcion `gorm:"foreignKey:UsuarioID"`
}

func (Usuario) TableName() string {
	return "usuarios"
}
