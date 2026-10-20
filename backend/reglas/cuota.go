package reglas

import "errors"

var (
	ErrPlanInvalido         = errors.New("plan inexistente")
	ErrActividadesNegativas = errors.New("la cantidad de actividades no puede ser negativa")
)

// CuotaMensual calcula la cuota de un socio según su plan, las actividades
// extra, si es estudiante y su antigüedad en meses.
// (A propósito, sin tests: es el PR que demuestra el freno por cobertura.)
func CuotaMensual(plan string, actividades int, estudiante bool, meses int) (float64, error) {
	if actividades < 0 {
		return 0, ErrActividadesNegativas
	}

	var base float64
	switch plan {
	case "basico":
		base = 20000
	case "full":
		base = 35000
	case "premium":
		base = 50000
	default:
		return 0, ErrPlanInvalido
	}

	extra := 0
	if actividades > 3 {
		extra = (actividades - 3) * 4000
	}
	total := base + float64(extra)

	if estudiante {
		total = total * 0.8
	}
	if meses >= 12 {
		total = total * 0.9
	} else if meses >= 6 {
		total = total * 0.95
	}

	if meses == 0 {
		total = total + 5000 // matrícula para socios nuevos
	}
	if total > 60000 {
		total = 60000 // tope de cuota
	}
	return total, nil
}
