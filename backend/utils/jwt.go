package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtDuration = time.Hour * 24
)

var jwtSecret string

func init() {
	jwtSecret = os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secreto123" // fallback en local
	}
}

// GenerateJWT genera un token JWT con claims
func GenerateJWT(userID uint, role string) (string, error) {
	expirationTime := time.Now().Add(jwtDuration)

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expirationTime),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "backend",
		Subject:   "auth",
		ID:        fmt.Sprintf("%d", userID),
		Audience:  jwt.ClaimStrings{role},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}
	return tokenString, nil
}

// ValidateJWT devuelve userID y rol
func ValidateJWT(tokenString string) (uint, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("signing method unexpected")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, "", errors.New("token inválido o expirado")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, "", errors.New("error extrayendo claims")
	}

	if claims.ID == "" {
		return 0, "", errors.New("user_id no encontrado en token")
	}
	var userID uint
	_, err = fmt.Sscanf(claims.ID, "%d", &userID)
	if err != nil {
		return 0, "", errors.New("error convirtiendo user_id")
	}

	// rol
	var role string
	if len(claims.Audience) > 0 {
		role = claims.Audience[0]
	}

	return userID, role, nil
}

// AuthMiddleware
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token faltante o mal formado"})
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")
		userID, role, err := ValidateJWT(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}
		ctx.Set("userID", userID)
		ctx.Set("rol", role)
		ctx.Next()
	}
}

// AdminMiddleware
func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token faltante o mal formado"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		userID, role, err := ValidateJWT(tokenString)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido"})
			return
		}

		if role != "administrador" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "permiso denegado, requiere rol administrador"})
			return
		}

		ctx.Set("userID", userID)
		ctx.Set("rol", role)
		ctx.Next()
	}
}
