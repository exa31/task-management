package usermodule

import (
	"database/sql"
	"errors"
	"time"

	appErrors "golang/pkg/errors"

	"github.com/google/uuid"
)

// UserService defines the business logic for the user domain.
type UserService interface {
	GetAll() ([]UserResponse, error)
	GetByID(id string) (*UserResponse, error)
	Create(req *CreateUserRequest) (*UserResponse, error)
	Update(id string, req *UpdateUserRequest) (*UserResponse, error)
	Delete(id string) error
	HardDelete(id string) error
}

// userService implements UserService.
type userService struct {
	repo UserRepository
}

// NewUserService builds the user service.
func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetAll() ([]UserResponse, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, appErrors.DatabaseError("Failed to fetch users", err)
	}

	res := make([]UserResponse, 0, len(users))
	for i := range users {
		res = append(res, users[i].toResponse())
	}
	return res, nil
}

func (s *userService) GetByID(id string) (*UserResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, appErrors.BadRequest("Invalid user ID format")
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, appErrors.DatabaseError("Failed to fetch user", err)
	}
	if user == nil {
		return nil, appErrors.NotFound("User not found")
	}

	u := user.toResponse()
	return &u, nil
}

func (s *userService) Create(req *CreateUserRequest) (*UserResponse, error) {
	exists, err := s.repo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, appErrors.DatabaseError("Failed to check email existence", err)
	}
	if exists {
		return nil, appErrors.Conflict("Email already exists")
	}

	now := time.Now()
	user := &User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Phone:     sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, appErrors.ParseDatabaseError(err)
	}

	u := user.toResponse()
	return &u, nil
}

func (s *userService) Update(id string, req *UpdateUserRequest) (*UserResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, appErrors.BadRequest("Invalid user ID format")
	}

	exists, err := s.repo.ExistsByID(id)
	if err != nil {
		return nil, appErrors.DatabaseError("Failed to check user existence", err)
	}
	if !exists {
		return nil, appErrors.NotFound("User not found")
	}

	if req.Email != "" {
		existing, err := s.repo.GetByEmail(req.Email)
		if err != nil {
			return nil, appErrors.DatabaseError("Failed to check email existence", err)
		}
		if existing != nil && existing.ID != id {
			return nil, appErrors.Conflict("Email already exists")
		}
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	user, err := s.repo.Update(id, updates)
	if err != nil {
		return nil, appErrors.ParseDatabaseError(err)
	}

	u := user.toResponse()
	return &u, nil
}

func (s *userService) Delete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return appErrors.BadRequest("Invalid user ID format")
	}

	err := s.repo.Delete(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NotFound("User not found")
		}
		return appErrors.ParseDatabaseError(err)
	}
	return nil
}

func (s *userService) HardDelete(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return appErrors.BadRequest("Invalid user ID format")
	}

	err := s.repo.HardDelete(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NotFound("User not found")
		}
		return appErrors.ParseDatabaseError(err)
	}
	return nil
}

func (u *User) toResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone.String,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
