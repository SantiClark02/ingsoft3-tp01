package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"backend/dao"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB inicializa la conexión a la base de datos usando variables de entorno y ejecuta las migraciones y el seed inicial.
func InitDB() {
	// Leer configuración desde variables de entorno
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Si alguna variable no está definida, loggear y terminar
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Variables de entorno DB_HOST, DB_PORT, DB_USER, DB_PASSWORD y DB_NAME deben estar definidas")
	}

	// Construir el DSN
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando con la base de datos: ", err)
	}

	// Ejecutar migraciones
	migrate()
}

func migrate() {
	// AutoMigrate sobre todas las entidades actuales
	err := DB.AutoMigrate(
		&dao.Usuario{},
		&dao.Actividad{},
		&dao.Horario{},
		&dao.Inscripcion{},
	)
	if err != nil {
		log.Fatal("Error al migrar las tablas: ", err)
	}

	seedDB()
}

func seedDB() {
	// Crea usuario administrador por defecto si no existe
	const defaultAdminEmail = "admin@gimnasio.com"
	const defaultAdminPassword = "admin123"

	// Verificar si ya existe un administrador con ese email
	var existing dao.Usuario
	if err := DB.Where("email = ?", defaultAdminEmail).First(&existing).Error; err == nil {
		// Ya existe, no hacemos nada
		return
	}

	// Hash de la contraseña
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Error al hashear la contraseña del administrador: %v", err)
	}

	admin := dao.Usuario{
		Username:     "admin", // Nombre de usuario por defecto
		Email:        defaultAdminEmail,
		PasswordHash: string(hashedPwd),
		Rol:          "administrador",
		CreatedAt:    time.Now(),
	}

	if err := DB.Create(&admin).Error; err != nil {
		log.Fatalf("Error al crear usuario administrador: %v", err)
	}
}
