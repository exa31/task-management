package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang/database"

	"github.com/jmoiron/sqlx"
)

type TokenRepository interface {
	CreateToken(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error)
	ListTokens(ctx context.Context, userID string) ([]APITokenResponse, error)
	DeleteToken(ctx context.Context, userID string, tokenID int64) error
	ValidateToken(ctx context.Context, rawToken string) (string, error)
}

type tokenRepo struct {
	db *sqlx.DB
}

func NewTokenRepository(db *sqlx.DB) TokenRepository {
	return &tokenRepo{db: db}
}

func (r *tokenRepo) exec(ctx context.Context) database.QueryExecutor {
	return database.GetExecutor(ctx, r.db)
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func generateRandomToken() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "mcp_pat_" + hex.EncodeToString(bytes), nil
}

func (r *tokenRepo) CreateToken(ctx context.Context, userID string, name string) (*GenerateTokenResponse, error) {
	rawToken, err := generateRandomToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := hashToken(rawToken)
	tokenPrefix := rawToken[:16]

	query := `
		INSERT INTO api_tokens (user_id, name, token_hash, token_prefix, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at, expires_at
	`
	var id int64
	var createdAt time.Time
	var expiresAt *time.Time
	err = r.exec(ctx).QueryRowxContext(ctx, query, userID, name, tokenHash, tokenPrefix).Scan(&id, &createdAt, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return &GenerateTokenResponse{
		ID:          id,
		Name:        name,
		TokenPrefix: tokenPrefix,
		RawToken:    rawToken,
		CreatedAt:   createdAt,
		ExpiresAt:   expiresAt,
	}, nil
}

func (r *tokenRepo) ListTokens(ctx context.Context, userID string) ([]APITokenResponse, error) {
	query := `
		SELECT id, user_id, name, token_prefix, last_used_at, created_at, expires_at
		FROM api_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	var tokens []APITokenResponse
	err := r.exec(ctx).SelectContext(ctx, &tokens, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tokens: %w", err)
	}
	if tokens == nil {
		tokens = []APITokenResponse{}
	}
	return tokens, nil
}

func (r *tokenRepo) DeleteToken(ctx context.Context, userID string, tokenID int64) error {
	query := `DELETE FROM api_tokens WHERE id = $1 AND user_id = $2`
	_, err := r.exec(ctx).ExecContext(ctx, query, tokenID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

func (r *tokenRepo) ValidateToken(ctx context.Context, rawToken string) (string, error) {
	tokenHash := hashToken(rawToken)
	query := `
		UPDATE api_tokens
		SET last_used_at = NOW()
		WHERE token_hash = $1
		RETURNING user_id
	`
	var userID string
	err := r.exec(ctx).QueryRowxContext(ctx, query, tokenHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("invalid or expired token")
		}
		return "", fmt.Errorf("failed to validate token: %w", err)
	}
	return userID, nil
}
