package services

import (
	"backend/clients"
	"backend/dao"
	"backend/utils"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Login valida credenciales y retorna el ID del usuario, un JWT y el rol si son correctas.
func Login(username string, password string) (uint, string, string, error) {
	var user dao.Usuario

	fmt.Println("➡️ Intentando login con usuario:", username)

	// Buscar usuario por nombre
	if tx := clients.DB.Where("nombre = ?", username).First(&user); tx.Error != nil {
		fmt.Println("❌ Usuario no encontrado en la base de datos")
		return 0, "", "", fmt.Errorf("usuario no encontrado")
	}

	// TP9 (CodeQL: "Clear-text logging of sensitive information", issue #14 del TP3):
	// nunca se loguean la contraseña ingresada ni el hash. Los logs los lee cualquiera
	// con acceso a docker logs / al runner, y no están cifrados.
	fmt.Println("➡️ Usuario encontrado:", user.Username)

	// Comparamos la contraseña ingresada con el hash almacenado
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		fmt.Println("❌ Falló la comparación de bcrypt:", err)
		return 0, "", "", errors.New("credenciales inválidas")
	}

	fmt.Println("✅ Contraseña válida, generando token...")

	// Generamos JWT usando utils.GenerateJWT
	token, err := utils.GenerateJWT(user.ID, user.Rol)
	if err != nil {
		fmt.Println("❌ Error generando token JWT:", err)
		return 0, "", "", fmt.Errorf("error generando token: %w", err)
	}

	fmt.Println("✅ Login exitoso: ID:", user.ID, "| Rol:", user.Rol)

	return user.ID, token, user.Rol, nil
}
