package main

import (
	"backend/controllers"
	database "backend/db"
	"backend/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	// Inicializar la base de datos (migraciones + seed)
	database.InitDB()

	// Inicializamos el router de Gin
	router := gin.New()

	// Middleware global: CORS y recuperación de pánico
	router.Use(utils.CORS, gin.Recovery())

	// Endpoint de vida para el smoke test del pipeline (TP6).
	// Devuelve también el commit que está corriendo (APP_COMMIT lo carga el deploy),
	// así el smoke puede comprobar que el entorno corre la versión que se desplegó.
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "commit": os.Getenv("APP_COMMIT")})
	})

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
