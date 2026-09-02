package services

import (
	"backend/clients"
	"backend/dao"
	"backend/domain"
	"fmt"
	"time"
)

// GetActivityByID busca en la base de datos una actividad por su ID.
func GetActivityByID(id int) (domain.Activity, error) {
	var daoAct dao.Actividad
	tx := clients.DB.First(&daoAct, id)
	if tx.Error != nil {
		return domain.Activity{}, fmt.Errorf("actividad con ID %d no encontrada: %w", id, tx.Error)
	}

	var daoHorarios []dao.Horario
	tx = clients.DB.
		Where("actividad_id = ?", id).
		Order("FIELD(dia_semana, 'Lun','Mar','Mie','Jue','Vie','Sab','Dom'), hora_inicio ASC").
		Find(&daoHorarios)
	if tx.Error != nil {
		return domain.Activity{}, fmt.Errorf("error cargando horarios para actividad %d: %w", id, tx.Error)
	}

	var schedule []domain.Horario
	for _, h := range daoHorarios {
		schedule = append(schedule, domain.Horario{
			ID:          h.ID,
			ActividadID: h.ActividadID,
			Day:         h.DiaSemana,
			StartTime:   h.HoraInicio.Format("15:04"),
			DurationMin: h.DuracionMin,
		})
	}

	domainAct := domain.Activity{
		ID:                  daoAct.ID,
		Name:                daoAct.Titulo,
		Description:         daoAct.Descripcion,
		Capacity:            daoAct.CupoMax,
		ImageURL:            daoAct.ImagenURL,
		CategoryName:        daoAct.NombreCategoria,
		InstructorName:      daoAct.NombreInstructor,
		InstructorEmail:     daoAct.EmailInstructor,
		InstructorSpecialty: daoAct.EspecialidadInstructor,
		Active:              daoAct.Activo,
		Schedule:            schedule,
	}

	return domainAct, nil
}

// ListarActividadesParaUsuario filtra actividades y marca las inscripciones de un usuario
func ListarActividadesParaUsuario(userID uint, query string) ([]domain.ActivityListItem, error) {
	inscRecs := []struct {
		InscriptionID uint
		ActividadID   uint
	}{}
	if userID != 0 {
		if err := clients.DB.
			Table("inscripciones").
			Select("inscripciones.id AS inscription_id, horarios.actividad_id").
			Joins("JOIN horarios ON inscripciones.horario_id = horarios.id").
			Where("inscripciones.usuario_id = ?", userID).
			Scan(&inscRecs).Error; err != nil {
			return nil, fmt.Errorf("error consultando inscripciones: %w", err)
		}
	}

	subscribedMap := make(map[uint]uint)
	for _, rec := range inscRecs {
		subscribedMap[rec.ActividadID] = rec.InscriptionID
	}

	var daoActs []dao.Actividad
	db := clients.DB
	if query != "" {
		like := "%" + query + "%"
		db = db.Where("titulo LIKE ? OR nombre_categoria LIKE ?", like, like)
	}
	if err := db.Find(&daoActs).Error; err != nil {
		return nil, fmt.Errorf("error listando actividades: %w", err)
	}

	var out []domain.ActivityListItem
	for _, a := range daoActs {
		var h dao.Horario
		clients.DB.
			Where("actividad_id = ?", a.ID).
			Order("FIELD(dia_semana,'Lun','Mar','Mie','Jue','Vie','Sab','Dom'), hora_inicio ASC").
			First(&h)

		item := domain.ActivityListItem{
			ID:             a.ID,
			Title:          a.Titulo,
			CategoryName:   a.NombreCategoria,
			InstructorName: a.NombreInstructor,
			Day:            h.DiaSemana,
			StartTime:      h.HoraInicio.Format("15:04"),
			Duration:       h.DuracionMin,
			Capacity:       a.CupoMax,
			ImageURL:       a.ImagenURL,
			Subscribed:     subscribedMap[a.ID] != 0,
			InscriptionID:  subscribedMap[a.ID],
		}
		out = append(out, item)
	}

	return out, nil
}

// ------------------------------------------------------------
// CREAR
// ------------------------------------------------------------
type NuevaActividadRequest struct {
	Titulo                 string              `json:"titulo"`
	Descripcion            string              `json:"descripcion"`
	CupoMax                int                 `json:"cupo_max"`
	ImagenURL              string              `json:"imagen_url"`
	NombreCategoria        string              `json:"nombre_categoria"`
	NombreInstructor       string              `json:"nombre_instructor"`
	EmailInstructor        string              `json:"email_instructor"`
	EspecialidadInstructor string              `json:"especialidad_instructor"`
	Activo                 bool                `json:"activo"`
	Horario                NuevoHorarioRequest `json:"horario"`
}

type NuevoHorarioRequest struct {
	ID          uint   `json:"id"`
	DiaSemana   string `json:"dia_semana"`
	HoraInicio  string `json:"hora_inicio"`
	DuracionMin int    `json:"duracion_min"`
}

func CreateActivity(req NuevaActividadRequest) error {
	actividad := dao.Actividad{
		Titulo:                 req.Titulo,
		Descripcion:            req.Descripcion,
		CupoMax:                uint16(req.CupoMax),
		ImagenURL:              &req.ImagenURL,
		NombreCategoria:        req.NombreCategoria,
		NombreInstructor:       req.NombreInstructor,
		EmailInstructor:        &req.EmailInstructor,
		EspecialidadInstructor: &req.EspecialidadInstructor,
		Activo:                 req.Activo,
	}

	if err := clients.DB.Create(&actividad).Error; err != nil {
		return fmt.Errorf("error creando actividad: %w", err)
	}

	horario := dao.Horario{
		ActividadID: actividad.ID,
		DiaSemana:   req.Horario.DiaSemana,
		HoraInicio:  parseHora(req.Horario.HoraInicio),
		DuracionMin: uint16(req.Horario.DuracionMin),
	}
	if err := clients.DB.Create(&horario).Error; err != nil {
		return fmt.Errorf("error creando horario: %w", err)
	}

	return nil
}

// ------------------------------------------------------------
// UPDATE
// ------------------------------------------------------------
func UpdateActivity(id int, req NuevaActividadRequest) error {
	var actividad dao.Actividad
	if err := clients.DB.First(&actividad, id).Error; err != nil {
		return fmt.Errorf("actividad no encontrada: %w", err)
	}

	// Actualizar campos de la actividad
	actividad.Titulo = req.Titulo
	actividad.Descripcion = req.Descripcion
	actividad.CupoMax = uint16(req.CupoMax)
	actividad.ImagenURL = &req.ImagenURL
	actividad.NombreCategoria = req.NombreCategoria
	actividad.NombreInstructor = req.NombreInstructor
	actividad.EmailInstructor = &req.EmailInstructor
	actividad.EspecialidadInstructor = &req.EspecialidadInstructor
	actividad.Activo = req.Activo

	if err := clients.DB.Save(&actividad).Error; err != nil {
		return fmt.Errorf("error actualizando actividad: %w", err)
	}

	// Intentar obtener horario existente
	var horario dao.Horario
	if err := clients.DB.Where("actividad_id = ?", actividad.ID).First(&horario).Error; err != nil {
		// no lo encontró => lo creamos nuevo
		newHorario := dao.Horario{
			ActividadID: actividad.ID,
			DiaSemana:   req.Horario.DiaSemana,
			HoraInicio:  parseHora(req.Horario.HoraInicio),
			DuracionMin: uint16(req.Horario.DuracionMin),
		}
		if err := clients.DB.Create(&newHorario).Error; err != nil {
			return fmt.Errorf("error creando horario nuevo: %w", err)
		}
	} else {
		// lo encontró => actualizar
		horario.DiaSemana = req.Horario.DiaSemana
		horario.HoraInicio = parseHora(req.Horario.HoraInicio)
		horario.DuracionMin = uint16(req.Horario.DuracionMin)
		if err := clients.DB.Save(&horario).Error; err != nil {
			return fmt.Errorf("error actualizando horario: %w", err)
		}
	}

	return nil
}

// ------------------------------------------------------------
// DELETE
// ------------------------------------------------------------
func DeleteActivity(id int) error {
	// primero borramos horarios
	if err := clients.DB.Where("actividad_id = ?", id).Delete(&dao.Horario{}).Error; err != nil {
		return fmt.Errorf("error eliminando horarios de la actividad: %w", err)
	}

	// luego la actividad
	if err := clients.DB.Delete(&dao.Actividad{}, id).Error; err != nil {
		return fmt.Errorf("error eliminando actividad: %w", err)
	}

	return nil
}

// parseHora convierte hh:mm a time.Time con año válido
func parseHora(horaStr string) time.Time {
	t, err := time.Parse("15:04", horaStr)
	if err != nil {
		return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(2000, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
}
