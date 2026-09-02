package controllers

import (
	"net/http"
	"strconv"

	"backend/services"

	"github.com/gin-gonic/gin"
)

// GetHorariosPorActividad maneja GET /activities/:id/horarios y devuelve todos los horarios asociados a una actividad.
func GetHorariosPorActividad(ctx *gin.Context) {
	// 1. Parsear ID de actividad
	idStr := ctx.Param("id")
	actID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de actividad inválido"})
		return
	}

	// Llama al servicio para listar los horarios
	horarios, err := services.ListarHorariosPorActividad(uint(actID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Devuelve JSON con la lista de horarios
	ctx.JSON(http.StatusOK, horarios)
}

// GetHorarioByID maneja GET /horarios/:id y devuelve un único horario por su ID.
func GetHorarioByID(ctx *gin.Context) {
	//Parsear ID de horario
	idStr := ctx.Param("id")
	horID, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID de horario inválido"})
		return
	}

	// Llama al servicio para obtener el horario
	horario, err := services.GetHorarioByID(uint(horID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	//Devulve JSON con el horario
	ctx.JSON(http.StatusOK, horario)
}
