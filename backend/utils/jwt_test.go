package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ---------- Regla 3: autenticación (JWT) ----------

func TestJWT_GenerarYValidar_DevuelveUsuarioYRol(t *testing.T) {
	token, _ := GenerateJWT(7, "socio")

	userID, rol, err := ValidateJWT(token)

	if err != nil || userID != 7 || rol != "socio" {
		t.Fatalf("esperaba (7, socio, nil), obtuve (%d, %s, %v)", userID, rol, err)
	}
}

// Caso de error: un token firmado con otro secreto no se acepta.
func TestValidarJWT_FirmadoConOtroSecreto_EsRechazado(t *testing.T) {
	falso, _ := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{ID: "1", Audience: jwt.ClaimStrings{"administrador"}}).
		SignedString([]byte("otro-secreto"))

	_, _, err := ValidateJWT(falso)

	if err == nil {
		t.Fatal("un token falsificado no debería validarse")
	}
}

func TestValidarJWT_TokenMalFormado_EsRechazado(t *testing.T) {
	_, _, err := ValidateJWT("esto-no-es-un-jwt")

	if err == nil {
		t.Fatal("un texto cualquiera no debería validarse como token")
	}
}

// ---------- Regla 4: autorización (sólo administradores) ----------

// Parametrizado: el mismo middleware con distintos pedidos.
func TestAdminMiddleware_SoloDejaPasarAdministradores(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokenSocio, _ := GenerateJWT(1, "socio")
	tokenAdmin, _ := GenerateJWT(2, "administrador")

	casos := []struct {
		nombre        string
		header        string
		codigo        int
		debeAbortarse bool
	}{
		{"sin token", "", http.StatusUnauthorized, true},
		{"header sin Bearer", tokenAdmin, http.StatusUnauthorized, true},
		{"socio", "Bearer " + tokenSocio, http.StatusForbidden, true},
		{"administrador", "Bearer " + tokenAdmin, http.StatusOK, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/activities", nil)
			if c.header != "" {
				ctx.Request.Header.Set("Authorization", c.header)
			}

			AdminMiddleware()(ctx)

			if ctx.IsAborted() != c.debeAbortarse || w.Code != c.codigo {
				t.Fatalf("esperaba abortado=%v código=%d, obtuve abortado=%v código=%d",
					c.debeAbortarse, c.codigo, ctx.IsAborted(), w.Code)
			}
		})
	}
}
