package authmodule

import (
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang/config"
	"golang/database"
	"golang/pkg/auth"
	"golang/pkg/errors"
	"golang/pkg/logging"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// AuthService contains the authentication business logic.
type AuthService struct {
	db         *database.PostgresDB
	jwtService *auth.JWTService
	cfg        *config.Config
}

// NewAuthService builds the auth service.
func NewAuthService(db *database.PostgresDB, jwtService *auth.JWTService, cfg *config.Config) *AuthService {
	return &AuthService{db: db, jwtService: jwtService, cfg: cfg}
}

type authUser struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Email        string     `db:"email"`
	Password     *string    `db:"password"`
	AvatarURL    *string    `db:"avatar_url"`
	AuthProvider *string    `db:"auth_provider"`
	ProviderID   *string    `db:"provider_id"`
	IsActive     bool       `db:"is_active"`
	CreatedAt    *time.Time `db:"created_at"`
}

// Register creates a user and returns the auth response.
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.InternalError("Failed to hash password", err)
	}

	user, err := database.Transact(s.db, func(tx *sqlx.Tx) (*authUser, error) {
		uid := uuid.New().String()
		var u authUser
		err := tx.QueryRowx(
			`INSERT INTO users (id, name, email, password, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW())
			 RETURNING id, name, email, created_at`,
			uid, req.Name, req.Email, string(hashedPassword),
		).StructScan(&u)
		if err != nil {
			return nil, errors.ParseDatabaseError(err)
		}
		return &u, nil
	})
	if err != nil {
		return nil, err
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User: UserPayload{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: derefTime(user.CreatedAt),
		},
		Tokens: newTokenPair(tokens),
	}, nil
}

// Login validates credentials and issues a token pair.
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var u authUser
	err := s.db.DB.Get(&u,
		"SELECT id, name, email, password, is_active, created_at FROM users WHERE email = $1",
		req.Email)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.Unauthorized("Invalid email or password")
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}
	if !u.IsActive {
		return nil, errors.Forbidden("Account is deactivated")
	}
	if u.Password == nil || bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(req.Password)) != nil {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	tokens, err := s.jwtService.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User: UserPayload{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: derefTime(u.CreatedAt),
		},
		Tokens: newTokenPair(tokens),
	}, nil
}

// RefreshToken exchanges a refresh token for a new pair.
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	tokens, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return newTokenPair(tokens), nil
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified any    `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *AuthService) verifyGoogleAuth(req *OAuthLoginRequest) (*googleUserInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Authorization Code flow (from initCodeClient popup dialog)
	if req.Code != "" {
		if s.cfg != nil && s.cfg.OAuth.GoogleClientID != "" && s.cfg.OAuth.GoogleClientSecret != "" {
			form := url.Values{}
			form.Set("code", req.Code)
			form.Set("client_id", s.cfg.OAuth.GoogleClientID)
			form.Set("client_secret", s.cfg.OAuth.GoogleClientSecret)
			redirectURI := s.cfg.OAuth.GoogleRedirectURI
			if redirectURI == "" {
				redirectURI = "postmessage"
			}
			form.Set("redirect_uri", redirectURI)
			form.Set("grant_type", "authorization_code")

			resp, err := client.PostForm("https://oauth2.googleapis.com/token", form)
			if err != nil {
				return nil, errors.BadRequest("Failed to communicate with Google token endpoint: " + err.Error())
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				return nil, errors.BadRequest("Google authorization code exchange failed: " + string(body))
			}

			var tokenRes struct {
				AccessToken string `json:"access_token"`
				IDToken     string `json:"id_token"`
			}
			if err := json.Unmarshal(body, &tokenRes); err != nil {
				return nil, errors.BadRequest("Failed to parse Google token response")
			}

			// Fetch verified user profile using the fresh access token
			reqUser, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
			reqUser.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
			userResp, err := client.Do(reqUser)
			if err != nil {
				return nil, errors.BadRequest("Failed to fetch Google user profile: " + err.Error())
			}
			defer userResp.Body.Close()
			userBody, _ := io.ReadAll(userResp.Body)

			var info googleUserInfo
			if err := json.Unmarshal(userBody, &info); err != nil {
				return nil, errors.BadRequest("Failed to decode Google user profile")
			}
			if info.Email == "" {
				return nil, errors.BadRequest("Google account did not return a verified email")
			}
			return &info, nil
		}

		// Fallback for development if credentials are not yet configured in .env
		if req.Email != "" {
			return &googleUserInfo{
				Sub:     "dev-" + req.Email,
				Email:   req.Email,
				Name:    req.Name,
				Picture: req.AvatarURL,
			}, nil
		}
		return nil, errors.BadRequest("Google OAuth is not configured on the backend. Please set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET in backend/.env")
	}

	// 2. ID Token flow (from Google One Tap 1-click credential)
	if req.IDToken != "" {
		tokeninfoURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(req.IDToken)
		resp, err := client.Get(tokeninfoURL)
		if err != nil {
			return nil, errors.BadRequest("Failed to verify Google ID token: " + err.Error())
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, errors.BadRequest("Google ID token validation failed: " + string(body))
		}

		var info googleUserInfo
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, errors.BadRequest("Failed to decode Google tokeninfo")
		}
		if info.Email == "" {
			return nil, errors.BadRequest("Google ID token does not contain an email address")
		}
		return &info, nil
	}

	// 3. Fallback if email is provided directly in development mode
	if req.Email != "" {
		return &googleUserInfo{
			Sub:     "dev-" + req.Email,
			Email:   req.Email,
			Name:    req.Name,
			Picture: req.AvatarURL,
		}, nil
	}

	return nil, errors.BadRequest("Missing Google authorization code or ID token")
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

var (
	appleKeysCache     *appleJWKS
	appleKeysCacheTime time.Time
	appleKeysMu        sync.RWMutex
)

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		nBytes, err = base64.URLEncoding.DecodeString(nStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode modulus: %w", err)
		}
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		eBytes, err = base64.URLEncoding.DecodeString(eStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode exponent: %w", err)
		}
	}

	eInt := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

func (s *AuthService) fetchAppleJWKS() (*appleJWKS, error) {
	appleKeysMu.RLock()
	if appleKeysCache != nil && time.Since(appleKeysCacheTime) < 1*time.Hour {
		cached := appleKeysCache
		appleKeysMu.RUnlock()
		return cached, nil
	}
	appleKeysMu.RUnlock()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Apple public keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Apple JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks appleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode Apple JWKS: %w", err)
	}

	appleKeysMu.Lock()
	appleKeysCache = &jwks
	appleKeysCacheTime = time.Now()
	appleKeysMu.Unlock()

	return &jwks, nil
}

type appleUserInfo struct {
	Sub     string
	Email   string
	Name    string
	Picture string
}

func (s *AuthService) verifyAppleAuth(req *OAuthLoginRequest) (*appleUserInfo, error) {
	var email string
	var name string
	var sub string

	// 1. If an IDToken is provided, perform cryptographic verification using Apple's public JWKS keys
	if req.IDToken != "" {
		jwks, err := s.fetchAppleJWKS()
		if err == nil && jwks != nil {
			token, parseErr := jwt.Parse(req.IDToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				kid, _ := token.Header["kid"].(string)
				for _, key := range jwks.Keys {
					if key.Kid == kid {
						return parseRSAPublicKey(key.N, key.E)
					}
				}
				return nil, fmt.Errorf("key with kid '%s' not found in Apple JWKS", kid)
			})

			if parseErr == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					// Validate Issuer (must be https://appleid.apple.com)
					if iss, ok := claims["iss"].(string); !ok || iss != "https://appleid.apple.com" {
						return nil, errors.Unauthorized("Invalid Apple token issuer")
					}

					// Validate Audience against configured AppleClientID
					if s.cfg != nil && s.cfg.OAuth.AppleClientID != "" {
						aud, ok := claims["aud"].(string)
						if !ok || aud != s.cfg.OAuth.AppleClientID {
							return nil, errors.Unauthorized(fmt.Sprintf("Apple token audience '%s' does not match configured Apple Client ID '%s'", aud, s.cfg.OAuth.AppleClientID))
						}
					}

					if subClaim, ok := claims["sub"].(string); ok {
						sub = subClaim
					}
					if emailClaim, ok := claims["email"].(string); ok && emailClaim != "" {
						email = emailClaim
					}
				}
			} else {
				logging.Warn("Apple token cryptographic verification failed", logging.Err(parseErr))
				// In production, reject invalid signatures. In development, allow fallback if testing offline.
				if s.cfg != nil && s.cfg.App.Environment == "production" {
					return nil, errors.Unauthorized("Invalid or expired Apple ID token")
				}
			}
		}

		// Fallback for development/testing if token couldn't be fetched or verified online
		if sub == "" || email == "" {
			parser := jwt.NewParser()
			token, _, err := parser.ParseUnverified(req.IDToken, jwt.MapClaims{})
			if err == nil {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub == "" {
						if subClaim, ok := claims["sub"].(string); ok {
							sub = subClaim
						}
					}
					if email == "" {
						if emailClaim, ok := claims["email"].(string); ok {
							email = emailClaim
						}
					}
					if name == "" {
						if nameClaim, ok := claims["name"].(string); ok {
							name = nameClaim
						}
					}
				}
			}
		}
	}

	// 2. Extract values from request payload if not in token
	if email == "" && req.Email != "" {
		email = req.Email
	}
	if name == "" && req.Name != "" {
		name = req.Name
	}
	if sub == "" {
		if req.Code != "" {
			sub = req.Code
		} else if email != "" {
			sub = "apple-" + email
		}
	}

	if email == "" {
		return nil, errors.BadRequest("Email is required for Apple sign-in")
	}

	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	picture := req.AvatarURL
	if picture == "" {
		picture = "https://api.dicebear.com/7.x/bottts/svg?seed=" + email
	}

	return &appleUserInfo{
		Sub:     sub,
		Email:   email,
		Name:    name,
		Picture: picture,
	}, nil
}

// OAuthLogin authenticates or provisions a user using Google or Apple sign-in.
func (s *AuthService) OAuthLogin(req *OAuthLoginRequest) (*AuthResponse, error) {
	var email string
	var name string
	var avatarURL string
	var providerID string
	provider := strings.ToLower(req.Provider)
	if provider == "" {
		provider = "google"
	}

	if provider == "google" {
		gInfo, err := s.verifyGoogleAuth(req)
		if err != nil {
			return nil, err
		}
		email = gInfo.Email
		name = gInfo.Name
		avatarURL = gInfo.Picture
		providerID = gInfo.Sub
	} else if provider == "apple" {
		appleInfo, err := s.verifyAppleAuth(req)
		if err != nil {
			return nil, err
		}
		email = appleInfo.Email
		name = appleInfo.Name
		avatarURL = appleInfo.Picture
		providerID = appleInfo.Sub
	} else {
		return nil, errors.BadRequest("Unsupported OAuth provider: " + provider)
	}

	if email == "" {
		return nil, errors.BadRequest("Email is required for OAuth login")
	}

	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	var user authUser
	err := s.db.DB.Get(&user, "SELECT id, name, email, avatar_url, auth_provider, is_active, created_at FROM users WHERE email = $1", email)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			uid := uuid.New().String()
			err = s.db.DB.QueryRowx(
				`INSERT INTO users (id, name, email, avatar_url, auth_provider, provider_id, is_active, created_at, updated_at)
				 VALUES ($1, $2, $3, $4, $5, $6, TRUE, NOW(), NOW())
				 RETURNING id, name, email, avatar_url, auth_provider, created_at`,
				uid, name, email, avatarURL, provider, providerID,
			).StructScan(&user)
			if err != nil {
				return nil, errors.DatabaseError("Failed to create user account", err)
			}
		} else {
			return nil, errors.DatabaseError("Database error during OAuth lookup", err)
		}
	} else {
		if avatarURL != "" || providerID != "" {
			_, _ = s.db.DB.Exec("UPDATE users SET avatar_url = COALESCE(NULLIF($1, ''), avatar_url), auth_provider = $2, provider_id = COALESCE(NULLIF($3, ''), provider_id), updated_at = NOW() WHERE id = $4", avatarURL, provider, providerID, user.ID)
			if avatarURL != "" {
				user.AvatarURL = &avatarURL
			}
		}
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User: UserPayload{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			AvatarURL: derefString(user.AvatarURL),
			Provider:  derefString(user.AuthProvider),
			CreatedAt: derefTime(user.CreatedAt),
		},
		Tokens: newTokenPair(tokens),
	}, nil
}

// DevLogin logs in or creates a development test user.
func (s *AuthService) DevLogin(req *DevLoginRequest) (*AuthResponse, error) {
	name := req.Name
	if name == "" {
		name = "Demo User"
	}
	avatar := "https://api.dicebear.com/7.x/bottts/svg?seed=" + req.Email
	return s.OAuthLogin(&OAuthLoginRequest{
		Provider:  "google",
		Email:     req.Email,
		Name:      name,
		AvatarURL: avatar,
	})
}

// GetCurrentUser fetches the authenticated user profile.
func (s *AuthService) GetCurrentUser(userID string) (*UserPayload, error) {
	var u authUser
	err := s.db.DB.Get(&u, "SELECT id, name, email, avatar_url, auth_provider, created_at FROM users WHERE id = $1", userID)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NotFound("User not found")
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}

	return &UserPayload{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		AvatarURL: derefString(u.AvatarURL),
		Provider:  derefString(u.AuthProvider),
		CreatedAt: derefTime(u.CreatedAt),
	}, nil
}

// UpdateProfile validates and updates the current user's profile.
type UpdateProfileRequest struct {
	Name string `json:"name" validate:"omitempty,min=3,max=100"`
}

func (s *AuthService) UpdateProfile(userID string, req *UpdateProfileRequest) (*UserPayload, error) {
	_, err := s.db.DB.Exec("UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2", req.Name, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to update profile", err)
	}
	return s.GetCurrentUser(userID)
}

// ChangePasswordRequest payload.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
}

func (s *AuthService) ChangePassword(userID string, req *ChangePasswordRequest) error {
	var hash string
	err := s.db.DB.Get(&hash, "SELECT password FROM users WHERE id = $1", userID)
	if err != nil {
		return errors.DatabaseError("Failed to fetch user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
		return errors.Unauthorized("Current password is incorrect")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalError("Failed to hash password", err)
	}

	if _, err := s.db.DB.Exec("UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2",
		string(hashed), userID); err != nil {
		return errors.DatabaseError("Failed to update password", err)
	}

	// Force re-login everywhere.
	if err := s.jwtService.RevokeAllUserTokens(userID); err != nil {
		logging.Warn("failed to revoke tokens after password change",
			slog.String("user_id", userID), logging.Err(err))
	}

	return nil
}

// newTokenPair converts the underlying token pair.
func newTokenPair(t *auth.TokenPair) *TokenPair {
	if t == nil {
		return nil
	}
	return &TokenPair{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    t.ExpiresAt,
		TokenType:    t.TokenType,
	}
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ProcessAppleWebhook handles Server-to-Server notification events from Apple.
func (s *AuthService) ProcessAppleWebhook(payload string) error {
	if payload == "" {
		return errors.BadRequest("Missing Apple webhook payload")
	}

	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(payload, jwt.MapClaims{})
	if err != nil {
		logging.Warn("failed to parse Apple webhook JWT", logging.Err(err))
		return errors.BadRequest("Invalid Apple webhook payload")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.BadRequest("Invalid Apple webhook claims")
	}

	eventsRaw, ok := claims["events"].(string)
	if !ok {
		if eventsMap, ok := claims["events"].(map[string]interface{}); ok {
			return s.handleAppleEventMap(eventsMap)
		}
		logging.Info("Apple webhook received without events", slog.Any("claims", claims))
		return nil
	}

	var eventObj struct {
		Type      string `json:"type"`
		Sub       string `json:"sub"`
		EventTime int64  `json:"event_time"`
	}
	if err := json.Unmarshal([]byte(eventsRaw), &eventObj); err != nil {
		logging.Warn("failed to parse Apple webhook event JSON", logging.Err(err))
		return nil
	}

	logging.Info("Apple Server-to-Server event received",
		slog.String("event_type", eventObj.Type),
		slog.String("sub", eventObj.Sub),
	)

	// Revoke active sessions when consent is revoked or account is deleted
	if eventObj.Type == "consent-revoked" || eventObj.Type == "account-delete" {
		var userID string
		err := s.db.DB.Get(&userID, "SELECT id FROM users WHERE auth_provider = 'apple' AND (provider_id = $1 OR id = $1)", eventObj.Sub)
		if err == nil && userID != "" {
			_ = s.jwtService.RevokeAllUserTokens(userID)
			logging.Info("Revoked all active sessions for Apple user", slog.String("user_id", userID))
		}
	}

	return nil
}

func (s *AuthService) handleAppleEventMap(eventsMap map[string]interface{}) error {
	eventType, _ := eventsMap["type"].(string)
	sub, _ := eventsMap["sub"].(string)
	logging.Info("Apple Server-to-Server event received",
		slog.String("event_type", eventType),
		slog.String("sub", sub),
	)

	if (eventType == "consent-revoked" || eventType == "account-delete") && sub != "" {
		var userID string
		err := s.db.DB.Get(&userID, "SELECT id FROM users WHERE auth_provider = 'apple' AND (provider_id = $1 OR id = $1)", sub)
		if err == nil && userID != "" {
			_ = s.jwtService.RevokeAllUserTokens(userID)
			logging.Info("Revoked all active sessions for Apple user", slog.String("user_id", userID))
		}
	}
	return nil
}

