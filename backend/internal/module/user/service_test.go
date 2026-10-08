package usermodule

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	customErrors "golang/pkg/errors"

	"github.com/google/uuid"
)

var (
	u1 = uuid.New().String()
	u2 = uuid.New().String()
)

// fakeUserRepository implements UserRepository in memory.
type fakeUserRepository struct {
	users   map[string]*User
	byEmail map[string]string // email -> id
}

func newFakeRepo(seed ...*User) *fakeUserRepository {
	r := &fakeUserRepository{users: map[string]*User{}, byEmail: map[string]string{}}
	for _, u := range seed {
		r.users[u.ID] = u
		r.byEmail[u.Email] = u.ID
	}
	return r
}

func (f *fakeUserRepository) GetAll() ([]User, error) {
	out := make([]User, 0, len(f.users))
	for _, u := range f.users {
		if u.IsActive {
			out = append(out, *u)
		}
	}
	return out, nil
}

func (f *fakeUserRepository) GetByID(id string) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (f *fakeUserRepository) GetByEmail(email string) (*User, error) {
	if id, ok := f.byEmail[email]; ok {
		return f.users[id], nil
	}
	return nil, nil
}

func (f *fakeUserRepository) Create(user *User) error {
	if _, exists := f.byEmail[user.Email]; exists {
		return errors.New("duplicate email")
	}
	f.users[user.ID] = user
	f.byEmail[user.Email] = user.ID
	return nil
}

func (f *fakeUserRepository) Update(id string, updates map[string]interface{}) (*User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	if name, ok := updates["name"].(string); ok {
		u.Name = name
	}
	if email, ok := updates["email"].(string); ok {
		delete(f.byEmail, u.Email)
		u.Email = email
		f.byEmail[email] = id
	}
	if phone, ok := updates["phone"].(string); ok {
		u.Phone = sql.NullString{String: phone, Valid: phone != ""}
	}
	u.UpdatedAt = time.Now()
	return u, nil
}

func (f *fakeUserRepository) Delete(id string) error {
	u, ok := f.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	u.IsActive = false
	return nil
}

func (f *fakeUserRepository) HardDelete(id string) error {
	if _, ok := f.users[id]; !ok {
		return sql.ErrNoRows
	}
	delete(f.users, id)
	return nil
}

func (f *fakeUserRepository) ExistsByID(id string) (bool, error) {
	_, ok := f.users[id]
	return ok, nil
}

func (f *fakeUserRepository) ExistsByEmail(email string) (bool, error) {
	_, ok := f.byEmail[email]
	return ok, nil
}

func seedUser(id, email string) *User {
	return &User{
		ID:        id,
		Name:      "Jane",
		Email:     email,
		Phone:     sql.NullString{String: "0811223344", Valid: true},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func mustBeCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	appErr, ok := err.(*customErrors.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != want {
		t.Errorf("code = %q, want %q", appErr.Code, want)
	}
}

func TestServiceGetAll(t *testing.T) {
	svc := NewUserService(newFakeRepo(seedUser(u1, "a@x.com"), seedUser(u2, "b@x.com")))
	users, err := svc.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestServiceGetByIDNotFound(t *testing.T) {
	svc := NewUserService(newFakeRepo())
	_, err := svc.GetByID(uuid.New().String())
	mustBeCode(t, err, "NOT_FOUND")
}

func TestServiceGetByIDInvalidUUID(t *testing.T) {
	svc := NewUserService(newFakeRepo())
	_, err := svc.GetByID("not-a-uuid")
	mustBeCode(t, err, "BAD_REQUEST")
}

func TestServiceCreateDuplicateEmail(t *testing.T) {
	svc := NewUserService(newFakeRepo(seedUser(u1, "dupe@x.com")))
	_, err := svc.Create(&CreateUserRequest{Name: "X", Email: "dupe@x.com"})
	mustBeCode(t, err, "CONFLICT")
}

func TestServiceCreateSuccess(t *testing.T) {
	svc := NewUserService(newFakeRepo())
	resp, err := svc.Create(&CreateUserRequest{Name: "New", Email: "new@x.com", Phone: "081000"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.ID == "" || resp.Email != "new@x.com" {
		t.Errorf("unexpected response: %#v", resp)
	}
}

func TestServiceUpdateSuccess(t *testing.T) {
	repo := newFakeRepo(seedUser(u1, "a@x.com"))
	svc := NewUserService(repo)
	resp, err := svc.Update(u1, &UpdateUserRequest{Name: "Renamed"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if resp.Name != "Renamed" {
		t.Errorf("expected renamed, got %#v", resp)
	}
}

func TestServiceUpdateEmailConflict(t *testing.T) {
	repo := newFakeRepo(seedUser(u1, "a@x.com"), seedUser(u2, "taken@x.com"))
	repo.byEmail["taken@x.com"] = u2
	svc := NewUserService(repo)
	_, err := svc.Update(u1, &UpdateUserRequest{Email: "taken@x.com"})
	mustBeCode(t, err, "CONFLICT")
}

func TestServiceDeleteNotFound(t *testing.T) {
	svc := NewUserService(newFakeRepo())
	err := svc.Delete(uuid.New().String())
	mustBeCode(t, err, "NOT_FOUND")
}

func TestServiceHardDeleteSuccess(t *testing.T) {
	svc := NewUserService(newFakeRepo(seedUser(u1, "a@x.com")))
	if err := svc.HardDelete(u1); err != nil {
		t.Fatalf("HardDelete failed: %v", err)
	}
}
