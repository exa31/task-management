package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"golang/config"
	customErrors "golang/pkg/errors"
	"golang/pkg/logging"

	"github.com/jmoiron/sqlx"
	// Register the postgres driver via its init side effect.
	_ "github.com/lib/pq"
)

// Transactor provides an interface for executing operations within a database transaction.
type Transactor interface {
	WithTransaction(fn func(tx *sqlx.Tx) error) error
	WithTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error
	WithReadOnlyTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error
}

type txContextKey struct{}

// ContextWithTx associates an active database transaction with the context.
func ContextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	if tx == nil {
		return ctx
	}
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext retrieves an active database transaction from the context, if one exists.
func TxFromContext(ctx context.Context) *sqlx.Tx {
	if ctx == nil {
		return nil
	}
	if tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// QueryExecutor is an interface satisfied by both *sqlx.DB and *sqlx.Tx.
type QueryExecutor interface {
	sqlx.ExtContext
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// GetExecutor returns the active transaction from context if present, otherwise returns db.
func GetExecutor(ctx context.Context, db *sqlx.DB) QueryExecutor {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

type PostgresDB struct {
	DB *sqlx.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	dsn := cfg.DSN()

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings with safe limits preventing connection exhaustion
	maxOpen := cfg.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 25
	}
	maxIdle := cfg.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = 15 * time.Minute
	}
	connMaxIdleTime := cfg.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 5 * time.Minute
	}

	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)
	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logging.Info("connected to PostgreSQL",
		slog.String("host", cfg.Host),
		slog.String("dbname", cfg.DBName),
	)

	return &PostgresDB{DB: db}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

// HealthCheck checks if database connection is alive
func (p *PostgresDB) HealthCheck() error {
	return p.DB.Ping()
}

// GetDB returns the underlying sqlx.DB instance
func (p *PostgresDB) GetDB() *sqlx.DB {
	return p.DB
}

// WithTransaction executes a function within a database transaction
// Automatically handles BEGIN, COMMIT, and ROLLBACK
// If the function returns an error, the transaction is rolled back
func (p *PostgresDB) WithTransaction(fn func(tx *sqlx.Tx) error) error {
	// Start transaction
	tx, err := p.DB.Beginx()
	if err != nil {
		return customErrors.DatabaseError("Failed to begin transaction", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if r := recover(); r != nil {
			// Rollback on panic
			if rbErr := tx.Rollback(); rbErr != nil {
				logging.Error("postgres rollback failed after panic", logging.Err(rbErr))
			}
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	err = fn(tx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Error("postgres transaction rollback failed", logging.Err(rbErr))
			return customErrors.DatabaseError("Transaction rollback failed", rbErr)
		}

		// Return the original error (might be AppError or database error)
		if appErr, ok := err.(*customErrors.AppError); ok {
			return appErr
		}
		return customErrors.ParseDatabaseError(err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return customErrors.DatabaseError("Failed to commit transaction", err)
	}

	return nil
}

// WithTransactionCtx executes a function within a database transaction with context
// Supports context cancellation and timeout
func (p *PostgresDB) WithTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	// If already inside an active transaction, reuse it directly
	if existingTx := TxFromContext(ctx); existingTx != nil {
		return fn(existingTx)
	}

	// Start transaction with context
	tx, err := p.DB.BeginTxx(ctx, nil)
	if err != nil {
		return customErrors.DatabaseError("Failed to begin transaction", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logging.Error("postgres rollback failed after panic", logging.Err(rbErr))
			}
			panic(r)
		}
	}()

	// Check if context is already canceled
	select {
	case <-ctx.Done():
		_ = tx.Rollback()
		return ctx.Err()
	default:
	}

	// Execute the function
	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Error("postgres transaction rollback failed", logging.Err(rbErr))
			return customErrors.DatabaseError("Transaction rollback failed", rbErr)
		}

		if appErr, ok := err.(*customErrors.AppError); ok {
			return appErr
		}
		return customErrors.ParseDatabaseError(err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return customErrors.DatabaseError("Failed to commit transaction", err)
	}

	return nil
}

// WithReadOnlyTransactionCtx executes a function within a read-only database transaction with context
func (p *PostgresDB) WithReadOnlyTransactionCtx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	// If already inside an active transaction, reuse it directly
	if existingTx := TxFromContext(ctx); existingTx != nil {
		return fn(existingTx)
	}

	// Start read-only transaction with context
	tx, err := p.DB.BeginTxx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return customErrors.DatabaseError("Failed to begin read-only transaction", err)
	}

	// Ensure transaction is finalized
	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logging.Error("postgres rollback failed after panic", logging.Err(rbErr))
			}
			panic(r)
		}
	}()

	// Check if context is already canceled
	select {
	case <-ctx.Done():
		_ = tx.Rollback()
		return ctx.Err()
	default:
	}

	// Execute the function
	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			logging.Error("postgres read-only transaction rollback failed", logging.Err(rbErr))
			return customErrors.DatabaseError("Transaction rollback failed", rbErr)
		}

		if appErr, ok := err.(*customErrors.AppError); ok {
			return appErr
		}
		return customErrors.ParseDatabaseError(err)
	}

	// Commit the read-only transaction
	if err := tx.Commit(); err != nil {
		return customErrors.DatabaseError("Failed to commit read-only transaction", err)
	}

	return nil
}

// Transact is a convenience wrapper that returns the result of a transaction
// Use this when you need to return data from the transaction
func Transact[T any](db *PostgresDB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	var result T

	err := db.WithTransaction(func(tx *sqlx.Tx) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}

// TransactCtx is a convenience wrapper with context support
func TransactCtx[T any](ctx context.Context, db *PostgresDB, fn func(tx *sqlx.Tx) (T, error)) (T, error) {
	var result T

	err := db.WithTransactionCtx(ctx, func(tx *sqlx.Tx) error {
		var err error
		result, err = fn(tx)
		return err
	})

	return result, err
}
