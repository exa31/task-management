package authmodule

import "time"

// RegisterRequest payload.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// LoginRequest payload.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// OAuthLoginRequest handles Google & Apple sign-in payloads.
type OAuthLoginRequest struct {
	Provider  string `json:"provider" validate:"required,oneof=google apple"`
	IDToken   string `json:"id_token"`
	Code      string `json:"code"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// DevLoginRequest allows instant demo authentication in local development.
type DevLoginRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name"`
}

// RefreshTokenRequest payload (optional in body if supplied via HttpOnly cookie).
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// UserPayload is a user summary for auth responses.
type UserPayload struct {
	ID                        string    `json:"id"`
	Name                      string    `json:"name"`
	Email                     string    `json:"email"`
	AvatarURL                 string    `json:"avatar_url"`
	Provider                  string    `json:"provider"`
	GoogleBound               bool      `json:"google_bound"`
	AppleBound                bool      `json:"apple_bound"`
	HasPassword               bool      `json:"has_password"`
	Timezone                  string    `json:"timezone"`
	DueDateWarningThreshold   int       `json:"due_date_warning_threshold"`
	EmailNotificationsEnabled bool      `json:"email_notifications_enabled"`
	CreatedAt                 time.Time `json:"created_at"`
}

// TokenPair mirrors the underlying token pair.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// AuthResponse is returned by oauth/login.
type AuthResponse struct {
	User   UserPayload `json:"user"`
	Tokens *TokenPair  `json:"tokens"`
}

// BindOAuthRequest handles binding Google or Apple account to an authenticated user.
type BindOAuthRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google apple"`
	IDToken  string `json:"id_token"`
	Code     string `json:"code"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// UnbindOAuthRequest handles unbinding an OAuth provider.
type UnbindOAuthRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google apple"`
}

// DeleteAccountRequest handles permanent deletion confirmation.
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation" validate:"required"`
}

// AppleWebhookRequest is the payload sent by Apple Server-to-Server notifications.
type AppleWebhookRequest struct {
	Payload string `json:"payload" form:"payload"`
}
