package controllers

import (
	"net/http"

	"backend/services"

	"github.com/gin-gonic/gin"
)

// loginRequest es el binding que acepta el endpoint POST /users/login
type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// loginResponse es la estructura que devolvemos al frontend luego de un login exitoso.
type loginResponse struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
	Rol    string `json:"rol"` // ← agregado para devolver el rol
}

// Login maneja POST /users/login y retorna user_id + JWT si las credenciales son válidas.
func Login(ctx *gin.Context) {
	var payload loginRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username y password requeridos"})
		return
	}

	// 👇 Cambiamos para recibir también el rol
	// Importante: services.Login DEBE usar utils.GenerateJWT, de lo contrario falla la validación
	userID, token, rol, err := services.Login(payload.Username, payload.Password)
	if err != nil {
		// Si falla la autenticación (usuario no existe o contraseña equivocada)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "credenciales inválidas"})
		return
	}

	ctx.JSON(http.StatusOK, loginResponse{
		UserID: userID,
		Token:  token,
		Rol:    rol, // ← lo agregamos a la respuesta
	})
}
