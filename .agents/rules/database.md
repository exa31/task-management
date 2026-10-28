---
description: Database transaction and query conventions for Go Fiber backend
trigger: glob
globs: "backend/**/*.go"
---

# Database & Query Rules

All database interactions in the Go Fiber backend must strictly adhere to the following rules:

---

## 1. Clean Repository Architecture

Repositories must adhere to clean architecture without ad-hoc `DBTX` interfaces or `WithTx` wrappers:

```go
type Repository interface {
    Create(ctx context.Context, userID string, req CreateRequest) (*Response, error)
    FindByID(ctx context.Context, userID string, id int64) (*Response, error)
    FindAll(ctx context.Context, userID string, filter Filter) ([]Response, error)
    Update(ctx context.Context, userID string, id int64, req UpdateRequest) (*Response, error)
    Delete(ctx context.Context, userID string, id int64) error
}

type repository struct {
    db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
    return &repository{db: db}
}
```

- Repositories accept `*sqlx.DB` directly in `NewRepository(db *sqlx.DB)`.
- Do **NOT** create a `DBTX` interface or expose `WithTx(tx *sqlx.Tx) Repository` on repositories.
- Services interact with the `Repository` interface directly (`s.repo.FindByID`, `s.repo.Create`, etc.).

---

## 2. Transactions via Database Helper

When atomic multi-step mutations are required across operations, use the transaction helper provided in `backend/database/postgres.go` (`WithTransaction` or `WithTransactionCtx`):

- **Automatic Commit & Rollback**: Returning `nil` commits the transaction; returning an `error` automatically rolls back.
- **Panic Recovery**: Any panic inside the helper automatically triggers a rollback and re-panics safely.

---

## 3. Query Best Practices & Safety

1. **SQL Injection Prevention**:
   - Always use parameterized placeholders (`$1`, `$2`, etc.). Never concatenate or format raw values into SQL strings.

2. **Context Propagation**:
   - Always use Context-aware query methods: `GetContext`, `SelectContext`, `QueryRowxContext`, `ExecContext` to respect HTTP request timeouts and client cancellations.

3. **No Partial Writes & No Silent Failures**:
   - Never ignore errors during child record inserts or deletes.
   - Any statement failure must immediately abort and return the error.
