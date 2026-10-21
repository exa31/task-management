package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"golang/database"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, userID string, req *CreateProjectRequest) (*ProjectResponse, error)
	FindAll(ctx context.Context, tx *sqlx.Tx, userID string) ([]ProjectResponse, error)
	FindByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*ProjectResponse, error)
	Update(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req *UpdateProjectRequest) (*ProjectResponse, error)
	Delete(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) exec(tx *sqlx.Tx) database.QueryExecutor {
	if tx != nil {
		return tx
	}
	return r.db
}

func (r *repository) Create(ctx context.Context, tx *sqlx.Tx, userID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	color := req.Color
	if color == "" {
		color = "indigo"
	}
	query := `
		INSERT INTO projects (user_id, name, color, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, user_id, name, color, description, 0 as task_count, created_at, updated_at
	`
	var res ProjectResponse
	err := r.exec(tx).QueryRowxContext(ctx, query, userID, req.Name, color, req.Description).StructScan(&res)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return &res, nil
}

func (r *repository) FindAll(ctx context.Context, tx *sqlx.Tx, userID string) ([]ProjectResponse, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.name, p.color, p.description, p.created_at, p.updated_at,
			COUNT(t.id) as task_count
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id AND t.user_id = p.user_id
		WHERE p.user_id = $1
		GROUP BY p.id
		ORDER BY p.name ASC
	`
	var items []ProjectResponse
	err := r.exec(tx).SelectContext(ctx, &items, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	if items == nil {
		items = []ProjectResponse{}
	}
	return items, nil
}

func (r *repository) FindByID(ctx context.Context, tx *sqlx.Tx, userID string, id int64) (*ProjectResponse, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.name, p.color, p.description, p.created_at, p.updated_at,
			COUNT(t.id) as task_count
		FROM projects p
		LEFT JOIN tasks t ON t.project_id = p.id AND t.user_id = p.user_id
		WHERE p.id = $1 AND p.user_id = $2
		GROUP BY p.id
	`
	var res ProjectResponse
	err := r.exec(tx).GetContext(ctx, &res, query, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find project: %w", err)
	}
	return &res, nil
}

func (r *repository) Update(ctx context.Context, tx *sqlx.Tx, userID string, id int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	query := "UPDATE projects SET updated_at = NOW()"
	args := []interface{}{id, userID}

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", len(args)+1)
		args = append(args, *req.Name)
	}
	if req.Color != nil {
		query += fmt.Sprintf(", color = $%d", len(args)+1)
		args = append(args, *req.Color)
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", len(args)+1)
		args = append(args, *req.Description)
	}

	query += " WHERE id = $1 AND user_id = $2 RETURNING id, user_id, name, color, description, 0 as task_count, created_at, updated_at"

	var res ProjectResponse
	err := r.exec(tx).QueryRowxContext(ctx, query, args...).StructScan(&res)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update project: %w", err)
	}
	return r.FindByID(ctx, tx, userID, id)
}

func (r *repository) Delete(ctx context.Context, tx *sqlx.Tx, userID string, id int64) error {
	query := `DELETE FROM projects WHERE id = $1 AND user_id = $2`
	result, err := r.exec(tx).ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
