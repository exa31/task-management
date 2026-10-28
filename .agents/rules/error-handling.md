---
description: Mandatory semantic error codes, English default messages, and frontend i18n locale synchronization
trigger: glob
globs: "{backend/**/*.go,frontend/src/i18n/locales/**/*.ts}"
---

# Error Handling & Localization Rules

All backend error returns and frontend error representations must strictly adhere to the following rules:

---

## 1. Mandatory Semantic Error Codes on Backend Error Returns

In the Go Fiber backend, the frontend error resolution pipeline (`frontend/src/api/errors.ts`) maps errors directly to translation keys using `errors.${apiErr.code}`.

1. **Never return bare generic errors for domain conditions**:
   - ❌ **Forbidden**:
     ```go
     // Bad: code defaults to generic "BAD_REQUEST", frontend cannot localize the specific issue
     return nil, errors.BadRequest("Failed to obtain provider identifier")
     return nil, errors.Conflict("Account is already linked")
     ```
   - ✅ **Required**:
     ```go
     // Good: explicit semantic error code from pkg/constants
     return nil, errors.BadRequestWithCode("Failed to obtain provider identifier", constants.CodeOAuthIdentifierMissing)
     return nil, errors.ConflictWithCode("This account is already linked to another user", constants.CodeOAuthAccountAlreadyLinked)
     ```

2. **Use `*WithCode` or Domain-Specific Constructors**:
   - `errors.BadRequestWithCode(message, code)`
   - `errors.UnauthorizedWithCode(message, code)`
   - `errors.ForbiddenWithCode(message, code)`
   - `errors.ConflictWithCode(message, code)`
   - `errors.NotFoundWithCode(message, code)`
   - `errors.InternalErrorWithCode(message, code, err)`
   - Or dedicated helpers:
     - `errors.TaskNotFound()` (`TASK_NOT_FOUND`)
     - `errors.SubtaskNotFound()` (`SUBTASK_NOT_FOUND`)
     - `errors.NoteNotFound()` (`NOTE_NOT_FOUND`)
     - `errors.ProjectNotFound()` (`PROJECT_NOT_FOUND`)
     - `errors.UserNotFound()` (`USER_NOT_FOUND`)
     - `errors.NoteNotActionable()` (`NOTE_NOT_ACTIONABLE`)
     - `errors.InvalidCredentials()` (`INVALID_CREDENTIALS`)
     - `errors.AccountDeactivated()` (`ACCOUNT_DEACTIVATED`)

---

## 2. Default Error Messages in Backend Must Be in English

All error messages defined on the backend (the `Message` field of `AppError`, database errors, validation messages, and success responses) **must be written in English by default**:

- ❌ **Forbidden**:
  ```go
  return errors.BadRequestWithCode("Konfirmasi tidak valid. Harap ketik DELETE", constants.CodeAccountDeleteInvalidConfirmation)
  return errors.NotFoundWithCode("User tidak ditemukan", constants.CodeUserNotFound)
  ```
- ✅ **Required**:
  ```go
  return errors.BadRequestWithCode("Invalid confirmation. Please type DELETE to delete your account.", constants.CodeAccountDeleteInvalidConfirmation)
  return errors.NotFoundWithCode("User not found", constants.CodeUserNotFound)
  ```

Localization to other languages (such as Indonesian) is strictly handled on the client-side via the frontend translation dictionary (`frontend/src/i18n/locales/id/errors.ts`) using the semantic `code`.

---

## 3. Centralized Error Code Constants

All error codes must be declared as exported constants in `backend/pkg/constants/database.go`:

- Use prefix `Code...` and `SCREAMING_SNAKE_CASE` string value:
  ```go
  const (
      CodeOAuthIdentifierMissing           = "OAUTH_IDENTIFIER_MISSING"
      CodeOAuthAccountAlreadyLinked        = "OAUTH_ACCOUNT_ALREADY_LINKED"
      CodeOAuthLastAuthMethod              = "OAUTH_LAST_AUTH_METHOD"
      CodeAccountDeleteInvalidConfirmation = "ACCOUNT_DELETE_INVALID_CONFIRMATION"
      CodeAvatarFileRequired               = "AVATAR_FILE_REQUIRED"
      CodeAvatarFileTooLarge                = "AVATAR_FILE_TOO_LARGE"
      CodeStorageNotConfigured             = "STORAGE_NOT_CONFIGURED"
  )
  ```

---

## 4. Strict Frontend Locale (i18n) Synchronization

Whenever a new error code is introduced in the backend, corresponding translation strings **MUST** be added to both language files:

1. **English**: `frontend/src/i18n/locales/en/errors.ts`
   ```ts
   OAUTH_IDENTIFIER_MISSING: 'Failed to obtain identifier from authentication provider.',
   OAUTH_ACCOUNT_ALREADY_LINKED: 'This account is already linked to another user.',
   ```

2. **Indonesian**: `frontend/src/i18n/locales/id/errors.ts`
   ```ts
   OAUTH_IDENTIFIER_MISSING: 'Gagal mendapatkan data identifikasi dari penyedia autentikasi.',
   OAUTH_ACCOUNT_ALREADY_LINKED: 'Akun ini sudah ditautkan ke akun pengguna lain.',
   ```

---

## 5. Verification Checklist

Before completing any task that introduces or modifies error returns:
1. Every error return in the Go code uses a semantic code from `pkg/constants`.
2. All default messages on the backend are in English.
3. Both `id/errors.ts` and `en/errors.ts` contain the matching key.
4. Backend unit tests pass: `go test ./...`
5. Frontend build and lint pass: `npm run lint && npm run build`

