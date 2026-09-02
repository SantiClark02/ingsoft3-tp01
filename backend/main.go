package main

import (
	"backend/controllers"
	database "backend/db"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

func main() {

	// Inicializar la base de datos (migraciones + seed)
	database.InitDB()

	// Inicializamos el router de Gin
	router := gin.New()

	// Middleware global: CORS y recuperación de pánico
	router.Use(utils.CORS, gin.Recovery())

	// Rutas de autenticación
	router.POST("/users/login", controllers.Login)

	// Rutas de actividades
	router.GET("/activities", controllers.GetAllActivities)
	router.GET("/activities/:id", controllers.GetActividadByID)

	router.POST("/activities", utils.AdminMiddleware(), controllers.CreateActividad)
	router.PUT("/activities/:id", utils.AdminMiddleware(), controllers.EditarActividad)
	router.DELETE("/activities/:id", utils.AdminMiddleware(), controllers.EliminarActividad)

	// Rutas de inscripciones
	router.POST("/inscripciones", controllers.CrearInscripcion)

	router.GET("/usuarios/:usuario_id/inscripciones", controllers.GetInscripcionesPorUsuario)
	router.DELETE("/inscripciones/:id", controllers.DeleteInscripcion)

	// Inicia el servidor en el puerto por defecto (":8080")
	router.Run()
}
