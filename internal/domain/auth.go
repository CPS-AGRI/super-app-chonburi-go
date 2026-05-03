package domain

type User struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	LastName    string   `json:"lastName"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         User   `json:"user"`
}

type AuthUseCase interface {
	Login(email, password string) (string, string, User, error)
	RefreshToken(token string) (string, string, User, error)
}

type RefreshTokenRepository interface {
	Create(token *AdminRefreshToken) error
	GetByToken(token string) (*AdminRefreshToken, error)
	DeleteByToken(token string) error
	DeleteByAdminID(adminID string) error
}
