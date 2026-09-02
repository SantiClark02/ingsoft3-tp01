package clients

import (
	"backend/dao"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB *gorm.DB
)

func init() {
	// Lee variables de entorno para Docker
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST") // esto vendrá como "mysql"
	port := 3306                 // puerto dentro del contenedor
	database := os.Getenv("DB_NAME")

	dsnFormat := "%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&loc=Local"
	dsn := fmt.Sprintf(dsnFormat, user, password, host, port, database)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("error connecting to database: %w", err))
	}

	// Ejecutar AutoMigrate para TODAS las entidades del proyecto
	err = DB.AutoMigrate(
		&dao.Usuario{},
		&dao.Actividad{},
		&dao.Horario{},
		&dao.Inscripcion{},
	)
	if err != nil {
		panic(fmt.Errorf("error en AutoMigrate: %w", err))
	}
}

// GetUserByUsername busca un usuario por su nombre de usuario
func GetUserByUsername(username string) (dao.Usuario, error) {
	var user dao.Usuario
	txn := DB.Where("nombre = ?", username).First(&user)
	if txn.Error != nil {
		return dao.Usuario{}, fmt.Errorf("error al buscar usuario: %w", txn.Error)
	}
	return user, nil
}
