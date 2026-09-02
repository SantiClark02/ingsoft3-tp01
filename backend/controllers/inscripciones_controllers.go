// Archivo: controllers/inscripciones_controllers.go
package controllers

import (
	"backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InscripcionRequest se usa en POST /inscripciones
type InscripcionRequest struct {
	UsuarioID uint `json:"usuario_id"`
	HorarioID uint `json:"horario_id"`
}

// CrearInscripcion maneja POST /inscripciones
func CrearInscripcion(ctx *gin.Context) {
	var req InscripcionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if req.UsuarioID == 0 || req.HorarioID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "usuario_id y horario_id son obligatorios"})
		return
	}

	// Llamamos al servicio, que ahora devuelve la Inscripcion completa
	insc, err := services.CrearInscripcion(req.UsuarioID, req.HorarioID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Respondemos con Created, mensaje, inscripcion_id y horario_id
	ctx.JSON(http.StatusCreated, gin.H{
		"mensaje":        "Inscripción creada exitosamente",
		"inscripcion_id": insc.ID,
		"horario_id":     insc.HorarioID,
	})
}

// PostInscripcion maneja POST /inscripciones/:horario_id para usuarios autenticados
func PostInscripcion(ctx *gin.Context) {
	horarioIDStr := ctx.Param("horario_id")
	horarioID, err := strconv.Atoi(horarioIDStr)
	if err != nil || horarioID <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de horario inválido"})
		return
	}

	userAny, ok := ctx.Get("user_id")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}
	userID := userAny.(uint)

	// Igualmente aquí desestructuramos
	insc, err := services.CrearInscripcion(userID, uint(horarioID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"mensaje":        "Inscripción realizada con éxito",
		"inscripcion_id": insc.ID,
	})
}

// GetInscripcionesPorUsuario maneja GET /usuarios/:usuario_id/inscripciones
func GetInscripcionesPorUsuario(ctx *gin.Context) {
	idStr := ctx.Param("usuario_id")
	uid, err := strconv.Atoi(idStr)
	if err != nil || uid <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}
	insc, err := services.ListarInscripcionesPorUsuario(uint(uid))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, insc)
}

// DeleteInscripcion maneja DELETE /inscripciones/:id
func DeleteInscripcion(ctx *gin.Context) {
	idStr := ctx.Param("id")
	iid, err := strconv.Atoi(idStr)
	if err != nil || iid <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de inscripción inválido"})
		return
	}
	if err := services.EliminarInscripcion(uint(iid)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}
