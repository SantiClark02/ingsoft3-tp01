package domain

// User se usa para binding de login y, opcionalmente, para representar datos de usuario.
type User struct {
	ID       uint   `json:"id,omitempty"`       // ID del usuario
	Username string `json:"username"`           // Nombre de usuario para login.
	Password string `json:"password,omitempty"` // Contraseña en texto plano solo para binding
	Email    string `json:"email,omitempty"`    // Correo del usuario
	Role     string `json:"role,omitempty"`     // Rol del usuario ("socio" o "administrador")
}
