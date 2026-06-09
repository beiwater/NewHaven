package auth

// Player represents an authenticated user.
type Player struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash,omitempty"` // excluded from API responses but included in internal snapshots
	DisplayName  string `json:"display_name,omitempty"`
	Gender       string `json:"gender,omitempty"`
	Email        string `json:"email,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// LoginRequest is the DTO for POST /api/login.
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse is the DTO for successful login.
type LoginResponse struct {
	Token     string `json:"token"`
	PlayerID  int    `json:"player_id"`
	CompanyID int    `json:"company_id"`
	Username  string `json:"username"`
}

// RegisterRequest is the DTO for POST /api/register.
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name,omitempty"`
	Gender   string `json:"gender,omitempty"`
	Email    string `json:"email,omitempty"`
}
