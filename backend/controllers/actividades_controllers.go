package controllers

import (
	"backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetActividadByID maneja GET /activities/:id y devuelve la actividad con todos sus horarios.
func GetActividadByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	actividad, err := services.GetActivityByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, actividad)
}

// GetAllActivities maneja GET /activities y devuelve todas las actividades
// opcionalmente filtradas por el query "search".
func GetAllActivities(ctx *gin.Context) {
	q := ctx.Query("search")
	var uid uint
	if s := ctx.Query("user_id"); s != "" {
		if tmp, err := strconv.Atoi(s); err == nil {
			uid = uint(tmp)
		}
	}

	acts, err := services.ListarActividadesParaUsuario(uid, q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acts)
}

// CreateActividad maneja POST /activities y crea una nueva actividad con horario inicial
func CreateActividad(ctx *gin.Context) {
	var nuevaActividad services.NuevaActividadRequest

	if err := ctx.ShouldBindJSON(&nuevaActividad); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	if err := services.CreateActivity(nuevaActividad); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"mensaje": "Actividad creada correctamente"})
}

// EditarActividad maneja PUT /activities/:id
func EditarActividad(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var editData services.NuevaActividadRequest
	if err := ctx.ShouldBindJSON(&editData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	if err := services.UpdateActivity(id, editData); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"mensaje": "Actividad actualizada correctamente"})
}

// EliminarActividad maneja DELETE /activities/:id
func EliminarActividad(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	if err := services.DeleteActivity(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"mensaje": "Actividad eliminada correctamente"})
}
