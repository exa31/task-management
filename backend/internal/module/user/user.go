package usermodule

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// User is the database entity for the user domain.
type User struct {
	ID        string         `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	Email     string         `json:"email" db:"email"`
	Phone     sql.NullString `json:"phone" db:"phone"`
	Password  string         `json:"-" db:"password"`
	IsActive  bool           `json:"is_active" db:"is_active"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

// CreateUserRequest payload.
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=100"`
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"omitempty,min=10,max=15"`
}

// UpdateUserRequest payload.
type UpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty,min=3,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
	Phone string `json:"phone" validate:"omitempty,min=10,max=15"`
}

// UserResponse is the public representation of a user.
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository defines data access for the user domain.
type UserRepository interface {
	GetAll() ([]User, error)
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	Update(id string, updates map[string]interface{}) (*User, error)
	Delete(id string) error
	HardDelete(id string) error
	ExistsByID(id string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}

// userRepository implements UserRepository.
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository builds the user repository.
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}
