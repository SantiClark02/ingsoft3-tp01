package main

import (
	"backend/controllers"
	database "backend/db"
	"backend/utils"
	"log"
	"net/http"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func main() {

	// TP9: Sentry (errores, demoras y uso). El DSN llega por el entorno: sólo PROD lo tiene.
	// Sin SENTRY_DSN (local, tests, QA, preprod) el SDK queda apagado y la app anda igual.
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("SENTRY_ENVIRONMENT"),
		Release:          os.Getenv("APP_COMMIT"), // qué imagen generó cada error
		EnableTracing:    true,
		TracesSampleRate: 1.0, // medir todos los pedidos (app chica: entra en el plan gratis)
	}); err != nil {
		log.Printf("Sentry no se pudo iniciar: %v", err)
	}

	// Inicializar la base de datos (migraciones + seed)
	database.InitDB()

	// Inicializamos el router de Gin
	router := gin.New()

	// Middleware global: CORS, recuperación de pánico y Sentry (TP9).
	// Sentry va adentro de Recovery: registra el pánico y lo re-lanza (Repanic)
	// para que Recovery conteste el 500 como siempre.
	router.Use(utils.CORS, gin.Recovery(), sentrygin.New(sentrygin.Options{Repanic: true}))

	// Endpoint de vida para el smoke test del pipeline (TP6).
	// Devuelve también el commit que está corriendo (APP_COMMIT lo carga el deploy),
	// así el smoke puede comprobar que el entorno corre la versión que se desplegó.
	health := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok", "commit": os.Getenv("APP_COMMIT")})
	}
	router.GET("/health", health)
	// TP9: el monitor externo gratuito pregunta con HEAD, y Gin no lo contesta solo (daba 404).
	router.HEAD("/health", health)

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
