package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS es el middleware básico para habilitar CORS en todas las rutas.
func CORS(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Si es una preflight request, respondemos con 200 y abortamos
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusOK)
		return
	}
	ctx.Next()
}
