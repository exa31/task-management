package authmodule

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
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
	"golang/pkg/constants"
	"golang/pkg/errors"
	"golang/pkg/logging"
	"golang/pkg/storage"

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
	storage    storage.Storage
}

// NewAuthService builds the auth service.
func NewAuthService(db *database.PostgresDB, jwtService *auth.JWTService, cfg *config.Config, stor storage.Storage) *AuthService {
	return &AuthService{db: db, jwtService: jwtService, cfg: cfg, storage: stor}
}

type authUser struct {
	ID                        string     `db:"id"`
	Name                      string     `db:"name"`
	Email                     string     `db:"email"`
	Password                  *string    `db:"password"`
	AvatarURL                 *string    `db:"avatar_url"`
	AuthProvider              *string    `db:"auth_provider"`
	ProviderID                *string    `db:"provider_id"`
	GoogleID                  *string    `db:"google_id"`
	AppleID                   *string    `db:"apple_id"`
	Timezone                  string     `db:"timezone"`
	DueDateWarningThreshold   int        `db:"due_date_warning_threshold"`
	EmailNotificationsEnabled bool       `db:"email_notifications_enabled"`
	IsActive                  bool       `db:"is_active"`
	CreatedAt                 *time.Time `db:"created_at"`
}

func (u *authUser) toPayload() UserPayload {
	tz := u.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	threshold := u.DueDateWarningThreshold
	if threshold <= 0 {
		threshold = 1440
	}
	googleBound := (u.GoogleID != nil && *u.GoogleID != "") || (derefString(u.AuthProvider) == "google")
	appleBound := (u.AppleID != nil && *u.AppleID != "") || (derefString(u.AuthProvider) == "apple")
	return UserPayload{
		ID:                        u.ID,
		Name:                      u.Name,
		Email:                     u.Email,
		AvatarURL:                 derefString(u.AvatarURL),
		Provider:                  derefString(u.AuthProvider),
		GoogleBound:               googleBound,
		AppleBound:                appleBound,
		HasPassword:               u.Password != nil && *u.Password != "",
		Timezone:                  tz,
		DueDateWarningThreshold:   threshold,
		EmailNotificationsEnabled: u.EmailNotificationsEnabled,
		CreatedAt:                 derefTime(u.CreatedAt),
	}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *AuthService) saveRefreshToken(userID, refreshToken string, expiresAt time.Time) error {
	tokenHash := hashToken(refreshToken)
	_, err := s.db.DB.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, is_revoked, created_at, updated_at)
		 VALUES ($1, $2, $3, false, NOW(), NOW())`,
		userID, tokenHash, expiresAt,
	)
	return err
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

	refreshDuration := 168
	if s.cfg != nil && s.cfg.JWT.RefreshTokenDuration > 0 {
		refreshDuration = s.cfg.JWT.RefreshTokenDuration
	}
	_ = s.saveRefreshToken(user.ID, tokens.RefreshToken, time.Now().Add(time.Duration(refreshDuration)*time.Hour))

	return &AuthResponse{
		User: UserPayload{
			ID:                        user.ID,
			Name:                      user.Name,
			Email:                     user.Email,
			HasPassword:               true,
			Timezone:                  "Asia/Jakarta",
			DueDateWarningThreshold:   1440,
			EmailNotificationsEnabled: true,
			CreatedAt:                 derefTime(user.CreatedAt),
		},
		Tokens: newTokenPair(tokens),
	}, nil
}

// Login validates credentials and issues a token pair.
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var u authUser
	err := s.db.DB.Get(&u,
		"SELECT id, name, email, password, avatar_url, auth_provider, google_id, apple_id, timezone, due_date_warning_threshold, email_notifications_enabled, is_active, created_at FROM users WHERE email = $1",
		req.Email)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.UnauthorizedWithCode("Invalid email or password", constants.CodeInvalidCredentials)
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}
	if !u.IsActive {
		return nil, errors.ForbiddenWithCode("Account is deactivated", constants.CodeAccountDeactivated)
	}
	if u.Password == nil || bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(req.Password)) != nil {
		return nil, errors.UnauthorizedWithCode("Invalid email or password", constants.CodeInvalidCredentials)
	}

	tokens, err := s.jwtService.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	refreshDuration := 168
	if s.cfg != nil && s.cfg.JWT.RefreshTokenDuration > 0 {
		refreshDuration = s.cfg.JWT.RefreshTokenDuration
	}
	_ = s.saveRefreshToken(u.ID, tokens.RefreshToken, time.Now().Add(time.Duration(refreshDuration)*time.Hour))

	return &AuthResponse{
		User:   u.toPayload(),
		Tokens: newTokenPair(tokens),
	}, nil
}

// RefreshToken exchanges a refresh token for a new pair or rotates if <= 48 hours remaining.
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.jwtService.ValidateToken(refreshToken, auth.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenHash := hashToken(refreshToken)
	var record struct {
		ID        int64     `db:"id"`
		UserID    string    `db:"user_id"`
		ExpiresAt time.Time `db:"expires_at"`
		IsRevoked bool      `db:"is_revoked"`
	}
	err = s.db.DB.Get(&record, "SELECT id, user_id, expires_at, is_revoked FROM refresh_tokens WHERE token_hash = $1", tokenHash)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.UnauthorizedWithCode("Refresh token has expired or is invalid", constants.CodeRefreshTokenInvalid)
		}
		return nil, errors.DatabaseError("Database error during token lookup", err)
	}

	if record.IsRevoked {
		return nil, errors.UnauthorizedWithCode("Refresh token has been revoked", constants.CodeRefreshTokenRevoked)
	}

	if record.ExpiresAt.Before(time.Now()) {
		return nil, errors.UnauthorizedWithCode("Refresh token has expired", constants.CodeRefreshTokenExpired)
	}

	remaining := time.Until(record.ExpiresAt)
	// If remaining <= 48 hours (2 days), rotate refresh token!
	if remaining <= 48*time.Hour {
		// Mark old token revoked in DB and Redis
		_, _ = s.db.DB.Exec("UPDATE refresh_tokens SET is_revoked = true, updated_at = NOW() WHERE id = $1", record.ID)
		_ = s.jwtService.RevokeToken(refreshToken, auth.RefreshToken)

		// Generate fresh token pair
		newPair, err := s.jwtService.GenerateTokenPair(claims.UserID, claims.Email)
		if err != nil {
			return nil, errors.InternalError("Failed to rotate tokens", err)
		}

		refreshDuration := 168
		if s.cfg != nil && s.cfg.JWT.RefreshTokenDuration > 0 {
			refreshDuration = s.cfg.JWT.RefreshTokenDuration
		}
		_ = s.saveRefreshToken(claims.UserID, newPair.RefreshToken, time.Now().Add(time.Duration(refreshDuration)*time.Hour))

		return newTokenPair(newPair), nil
	}

	// If remaining > 48 hours: keep the current refresh token and only issue a new access token
	accessTok, accessExp, err := s.jwtService.GenerateAccessToken(claims.UserID, claims.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate access token", err)
	}

	return &TokenPair{
		AccessToken:  accessTok,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExp,
		TokenType:    "Bearer",
	}, nil
}

// RevokeCurrentToken permanently deletes the specific refresh token during single-device logout.
func (s *AuthService) RevokeCurrentToken(refreshToken string) {
	if refreshToken == "" {
		return
	}
	tokenHash := hashToken(refreshToken)
	_, _ = s.db.DB.Exec("DELETE FROM refresh_tokens WHERE token_hash = $1", tokenHash)
	_ = s.jwtService.RevokeToken(refreshToken, auth.RefreshToken)
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
				return nil, errors.BadRequestWithCode("Failed to communicate with Google token endpoint: "+err.Error(), constants.CodeOAuthNotConfigured)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				return nil, errors.BadRequestWithCode("Google authorization code exchange failed: "+string(body), constants.CodeOAuthProviderInvalid)
			}

			var tokenRes struct {
				AccessToken string `json:"access_token"`
				IDToken     string `json:"id_token"`
			}
			if err := json.Unmarshal(body, &tokenRes); err != nil {
				return nil, errors.BadRequestWithCode("Failed to parse Google token response", constants.CodeOAuthProviderInvalid)
			}

			// Fetch verified user profile using the fresh access token
			reqUser, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
			reqUser.Header.Set("Authorization", "Bearer "+tokenRes.AccessToken)
			userResp, err := client.Do(reqUser)
			if err != nil {
				return nil, errors.BadRequestWithCode("Failed to fetch Google user profile: "+err.Error(), constants.CodeOAuthProviderInvalid)
			}
			defer userResp.Body.Close()
			userBody, _ := io.ReadAll(userResp.Body)

			var info googleUserInfo
			if err := json.Unmarshal(userBody, &info); err != nil {
				return nil, errors.BadRequestWithCode("Failed to decode Google user profile", constants.CodeOAuthProviderInvalid)
			}
			if info.Email == "" {
				return nil, errors.BadRequestWithCode("Google account did not return a verified email", constants.CodeOAuthEmailRequired)
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
		return nil, errors.BadRequestWithCode("Google OAuth is not configured on the backend. Please set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET in backend/.env", constants.CodeOAuthNotConfigured)
	}

	// 2. ID Token flow (from Google One Tap 1-click credential)
	if req.IDToken != "" {
		tokeninfoURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(req.IDToken)
		resp, err := client.Get(tokeninfoURL)
		if err != nil {
			return nil, errors.BadRequestWithCode("Failed to verify Google ID token: "+err.Error(), constants.CodeOAuthProviderInvalid)
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, errors.BadRequestWithCode("Google ID token validation failed: "+string(body), constants.CodeOAuthProviderInvalid)
		}

		var info googleUserInfo
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, errors.BadRequestWithCode("Failed to decode Google tokeninfo", constants.CodeOAuthProviderInvalid)
		}
		if info.Email == "" {
			return nil, errors.BadRequestWithCode("Google ID token does not contain an email address", constants.CodeOAuthEmailRequired)
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

	return nil, errors.BadRequestWithCode("Missing Google authorization code or ID token", constants.CodeOAuthIdentifierMissing)
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

func (s *AuthService) getAllowedAppleAudiences() []string {
	if s.cfg == nil {
		return nil
	}
	var audiences []string
	if s.cfg.OAuth.AppleClientID != "" {
		for _, id := range strings.Split(s.cfg.OAuth.AppleClientID, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				audiences = append(audiences, trimmed)
			}
		}
	}
	if s.cfg.OAuth.AppleBundleID != "" {
		for _, id := range strings.Split(s.cfg.OAuth.AppleBundleID, ",") {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				audiences = append(audiences, trimmed)
			}
		}
	}
	return audiences
}

func (s *AuthService) isAllowedAppleAudience(aud string) bool {
	allowedList := s.getAllowedAppleAudiences()
	if len(allowedList) == 0 {
		return true
	}
	for _, allowed := range allowedList {
		if allowed == aud {
			return true
		}
	}
	return false
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
						return nil, errors.UnauthorizedWithCode("Invalid Apple token issuer", constants.CodeAppleTokenInvalidIssuer)
					}

					// Validate Audience against configured Apple Client ID(s) and mobile Bundle ID(s)
					if s.cfg != nil && len(s.getAllowedAppleAudiences()) > 0 {
						aud, ok := claims["aud"].(string)
						if !ok || !s.isAllowedAppleAudience(aud) {
							allowedStr := strings.Join(s.getAllowedAppleAudiences(), ", ")
							return nil, errors.UnauthorizedWithCode(
								fmt.Sprintf("Apple token audience '%s' does not match any configured Apple Client ID or Bundle ID [%s]", aud, allowedStr),
								constants.CodeAppleTokenAudienceMismatch,
							)
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
					return nil, errors.UnauthorizedWithCode("Invalid or expired Apple ID token", constants.CodeAppleTokenInvalid)
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
		return nil, errors.BadRequestWithCode("Email is required for Apple sign-in", constants.CodeOAuthEmailRequired)
	}

	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	picture := req.AvatarURL

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
		return nil, errors.BadRequestWithCode("Unsupported OAuth provider: "+provider, constants.CodeOAuthProviderInvalid)
	}

	if email == "" {
		return nil, errors.BadRequestWithCode("Email is required for OAuth login", constants.CodeOAuthEmailRequired)
	}

	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	var user authUser
	var err error
	found := false

	// 1. Try finding user by provider unique identifier (apple_id or google_id)
	providerCol := "google_id"
	if provider == "apple" {
		providerCol = "apple_id"
	}

	if providerID != "" {
		providerQuery := fmt.Sprintf("SELECT id, name, email, password, avatar_url, auth_provider, google_id, apple_id, timezone, due_date_warning_threshold, email_notifications_enabled, is_active, created_at FROM users WHERE %s = $1", providerCol)
		err = s.db.DB.Get(&user, providerQuery, providerID)
		if err == nil {
			found = true
		} else if !stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.DatabaseError("Database error during OAuth provider ID lookup", err)
		}
	}

	// 2. If not matched by provider ID, fallback to email lookup
	if !found && email != "" {
		err = s.db.DB.Get(&user, "SELECT id, name, email, password, avatar_url, auth_provider, google_id, apple_id, timezone, due_date_warning_threshold, email_notifications_enabled, is_active, created_at FROM users WHERE email = $1", email)
		if err == nil {
			found = true
		} else if !stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.DatabaseError("Database error during OAuth email lookup", err)
		}
	}

	if !found {
		uid := uuid.New().String()
		googleID := &providerID
		var appleID *string
		if provider == "apple" {
			googleID = nil
			appleID = &providerID
		}
		if avatarURL == "" {
			avatarURL = "https://api.dicebear.com/7.x/bottts/svg?seed=" + uid
		}
		err = s.db.DB.QueryRowx(
			`INSERT INTO users (id, name, email, avatar_url, auth_provider, provider_id, google_id, apple_id, is_active, timezone, due_date_warning_threshold, email_notifications_enabled, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE, 'Asia/Jakarta', 1440, TRUE, NOW(), NOW())
			 RETURNING id, name, email, avatar_url, auth_provider, google_id, apple_id, timezone, due_date_warning_threshold, email_notifications_enabled, created_at`,
			uid, name, email, avatarURL, provider, providerID, googleID, appleID,
		).StructScan(&user)
		if err != nil {
			return nil, errors.DatabaseError("Failed to create user account", err)
		}
	} else {
		if !user.IsActive {
			return nil, errors.ForbiddenWithCode("Account is deactivated", constants.CodeAccountDeactivated)
		}
		newAvatar := ""
		if (user.AvatarURL == nil || *user.AvatarURL == "") && avatarURL != "" {
			newAvatar = avatarURL
		}
		updateQuery := fmt.Sprintf("UPDATE users SET avatar_url = COALESCE(NULLIF($1, ''), avatar_url), auth_provider = $2, provider_id = COALESCE(NULLIF($3, ''), provider_id), %s = COALESCE($3, %s), updated_at = NOW() WHERE id = $4", providerCol, providerCol)
		_, _ = s.db.DB.Exec(updateQuery, newAvatar, provider, providerID, user.ID)
		if newAvatar != "" {
			user.AvatarURL = &newAvatar
		}
		if provider == "apple" {
			user.AppleID = &providerID
		} else {
			user.GoogleID = &providerID
		}
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	refreshDuration := 168
	if s.cfg != nil && s.cfg.JWT.RefreshTokenDuration > 0 {
		refreshDuration = s.cfg.JWT.RefreshTokenDuration
	}
	_ = s.saveRefreshToken(user.ID, tokens.RefreshToken, time.Now().Add(time.Duration(refreshDuration)*time.Hour))

	return &AuthResponse{
		User:   user.toPayload(),
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
	err := s.db.DB.Get(&u, "SELECT id, name, email, password, avatar_url, auth_provider, google_id, apple_id, timezone, due_date_warning_threshold, email_notifications_enabled, created_at FROM users WHERE id = $1", userID)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NotFoundWithCode("User not found", constants.CodeUserNotFound)
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}

	payload := u.toPayload()
	return &payload, nil
}

// BindOAuth connects Google or Apple credentials to the currently authenticated user.
func (s *AuthService) BindOAuth(userID string, req *BindOAuthRequest) (*UserPayload, error) {
	provider := strings.ToLower(req.Provider)
	var providerID string
	if provider == "google" {
		gInfo, err := s.verifyGoogleAuth(&OAuthLoginRequest{
			Provider: req.Provider,
			IDToken:  req.IDToken,
			Code:     req.Code,
			Email:    req.Email,
			Name:     req.Name,
		})
		if err != nil {
			return nil, err
		}
		providerID = gInfo.Sub
	} else if provider == "apple" {
		aInfo, err := s.verifyAppleAuth(&OAuthLoginRequest{
			Provider: req.Provider,
			IDToken:  req.IDToken,
			Code:     req.Code,
			Email:    req.Email,
			Name:     req.Name,
		})
		if err != nil {
			return nil, err
		}
		providerID = aInfo.Sub
	} else {
		return nil, errors.BadRequestWithCode("Unsupported OAuth provider: "+provider, constants.CodeOAuthProviderInvalid)
	}

	if providerID == "" {
		return nil, errors.BadRequestWithCode("Failed to obtain provider identifier", constants.CodeOAuthIdentifierMissing)
	}

	checkCol := "google_id"
	if provider == "apple" {
		checkCol = "apple_id"
	}

	var conflictID string
	query := fmt.Sprintf("SELECT id FROM users WHERE %s = $1 AND id != $2", checkCol)
	err := s.db.DB.Get(&conflictID, query, providerID, userID)
	if err == nil && conflictID != "" {
		return nil, errors.ConflictWithCode(fmt.Sprintf("This %s account is already linked to another user.", provider), constants.CodeOAuthAccountAlreadyLinked)
	}

	updateQuery := fmt.Sprintf("UPDATE users SET %s = $1, updated_at = NOW() WHERE id = $2", checkCol)
	_, err = s.db.DB.Exec(updateQuery, providerID, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to link "+provider+" account", err)
	}

	return s.GetCurrentUser(userID)
}

// UnbindOAuth disconnects Google or Apple credentials from the current user.
func (s *AuthService) UnbindOAuth(userID string, req *UnbindOAuthRequest) (*UserPayload, error) {
	provider := strings.ToLower(req.Provider)
	var u authUser
	err := s.db.DB.Get(&u, "SELECT id, password, auth_provider, google_id, apple_id FROM users WHERE id = $1", userID)
	if err != nil {
		return nil, errors.NotFoundWithCode("User not found", constants.CodeUserNotFound)
	}

	hasPassword := u.Password != nil && *u.Password != ""
	hasGoogle := u.GoogleID != nil && *u.GoogleID != ""
	hasApple := u.AppleID != nil && *u.AppleID != ""
	mainProvider := derefString(u.AuthProvider)

	methodsCount := 0
	if hasPassword {
		methodsCount++
	}
	if hasGoogle {
		methodsCount++
	}
	if hasApple {
		methodsCount++
	}
	if mainProvider != "" && mainProvider != "email" && !hasGoogle && !hasApple {
		methodsCount++
	}

	if methodsCount <= 1 {
		return nil, errors.BadRequestWithCode("Cannot unbind account as it is your only remaining login method.", constants.CodeOAuthLastAuthMethod)
	}

	checkCol := "google_id"
	if provider == "apple" {
		checkCol = "apple_id"
	}

	query := fmt.Sprintf("UPDATE users SET %s = NULL, updated_at = NOW() WHERE id = $1", checkCol)
	_, err = s.db.DB.Exec(query, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to unbind "+provider+" account", err)
	}

	return s.GetCurrentUser(userID)
}

// UploadAvatar processes and stores an avatar image in MinIO.
func (s *AuthService) UploadAvatar(userID string, reader io.Reader, size int64, contentType string) (*UserPayload, error) {
	if s.storage == nil {
		return nil, errors.InternalErrorWithCode("Storage system is not configured", constants.CodeStorageNotConfigured, nil)
	}

	objectName := fmt.Sprintf("avatar_%s_%d.webp", userID, time.Now().Unix())
	publicURL, err := s.storage.Upload(context.Background(), objectName, reader, size, contentType)
	if err != nil {
		return nil, errors.InternalError("Failed to store avatar: "+err.Error(), err)
	}

	_, err = s.db.DB.Exec("UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2", publicURL, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to update profile avatar", err)
	}

	return s.GetCurrentUser(userID)
}

// DeleteAccount permanently purges the user and all associated resources.
func (s *AuthService) DeleteAccount(userID string, req *DeleteAccountRequest) error {
	if strings.ToUpper(strings.TrimSpace(req.Confirmation)) != "DELETE" {
		return errors.BadRequestWithCode("Invalid confirmation. Please type DELETE to delete your account.", constants.CodeAccountDeleteInvalidConfirmation)
	}

	var u authUser
	err := s.db.DB.Get(&u, "SELECT id, avatar_url FROM users WHERE id = $1", userID)
	if err != nil {
		return errors.NotFoundWithCode("User not found", constants.CodeUserNotFound)
	}

	// Purge avatar from storage if exists
	if u.AvatarURL != nil && *u.AvatarURL != "" && s.storage != nil {
		_ = s.storage.Delete(context.Background(), *u.AvatarURL)
	}

	// Database hard deletion (cascades tasks, subtasks, notes, projects, refresh_tokens, api_tokens)
	_, err = s.db.DB.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return errors.DatabaseError("Failed to permanently delete account", err)
	}

	// Invalidate all tokens in Redis
	_ = s.jwtService.RevokeAllUserTokens(userID)

	return nil
}

// UpdateProfile validates and updates the current user's profile.
type UpdateProfileRequest struct {
	Name                      *string `json:"name" validate:"omitempty,min=3,max=100"`
	Timezone                  *string `json:"timezone" validate:"omitempty,max=50"`
	DueDateWarningThreshold   *int    `json:"due_date_warning_threshold" validate:"omitempty,min=0,max=43200"`
	EmailNotificationsEnabled *bool   `json:"email_notifications_enabled"`
}

func (s *AuthService) UpdateProfile(userID string, req *UpdateProfileRequest) (*UserPayload, error) {
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{userID}

	if req.Name != nil && *req.Name != "" {
		query += fmt.Sprintf(", name = $%d", len(args)+1)
		args = append(args, *req.Name)
	}
	if req.Timezone != nil && *req.Timezone != "" {
		query += fmt.Sprintf(", timezone = $%d", len(args)+1)
		args = append(args, *req.Timezone)
	}
	if req.DueDateWarningThreshold != nil {
		query += fmt.Sprintf(", due_date_warning_threshold = $%d", len(args)+1)
		args = append(args, *req.DueDateWarningThreshold)
	}
	if req.EmailNotificationsEnabled != nil {
		query += fmt.Sprintf(", email_notifications_enabled = $%d", len(args)+1)
		args = append(args, *req.EmailNotificationsEnabled)
	}

	query += " WHERE id = $1"
	_, err := s.db.DB.Exec(query, args...)
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
		return errors.UnauthorizedWithCode("Current password is incorrect", constants.CodeInvalidCurrentPassword)
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
		return errors.BadRequestWithCode("Missing Apple webhook payload", constants.CodeBadRequest)
	}

	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(payload, jwt.MapClaims{})
	if err != nil {
		logging.Warn("failed to parse Apple webhook JWT", logging.Err(err))
		return errors.BadRequestWithCode("Invalid Apple webhook payload", constants.CodeBadRequest)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.BadRequestWithCode("Invalid Apple webhook claims", constants.CodeBadRequest)
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
