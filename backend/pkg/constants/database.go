package constants

// PostgreSQL constraint names mapped to field names
// This helps provide user-friendly error messages for database constraint violations
var UniqueConstraintFieldMap = map[string]string{
	"skills_name_key":      "name",
	"users_email_key":      "email",
	"users_username_key":   "username",
	"users_phone_key":      "phone",
	"products_sku_key":     "sku",
	"products_slug_key":    "slug",
	"categories_name_key":  "name",
	"categories_slug_key":  "slug",
	"roles_name_key":       "name",
	"permissions_name_key": "name",
}

// Foreign key constraint names mapped to readable names
var ForeignKeyConstraintMap = map[string]string{
	"users_role_id_fkey":                  "role",
	"orders_user_id_fkey":                 "user",
	"order_items_order_id_fkey":           "order",
	"order_items_product_id_fkey":         "product",
	"products_category_id_fkey":           "category",
	"user_permissions_user_id_fkey":       "user",
	"user_permissions_permission_id_fkey": "permission",
}

// Error codes
const (
	// Success codes
	CodeSuccess = "SUCCESS"
	CodeCreated = "CREATED"

	// Client error codes (4xx)
	CodeBadRequest          = "BAD_REQUEST"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeDuplicateEntry      = "DUPLICATE_ENTRY"
	CodeForeignKeyViolation = "FOREIGN_KEY_VIOLATION"
	CodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
	CodeRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"

	// Domain-specific error codes (4xx)
	CodeNoteNotFound                     = "NOTE_NOT_FOUND"
	CodeNoteNotActionable                = "NOTE_NOT_ACTIONABLE"
	CodeTaskNotFound                     = "TASK_NOT_FOUND"
	CodeSubtaskNotFound                  = "SUBTASK_NOT_FOUND"
	CodeProjectNotFound                  = "PROJECT_NOT_FOUND"
	CodeMcpTokenNotFound                 = "MCP_TOKEN_NOT_FOUND"
	CodeUserNotFound                     = "USER_NOT_FOUND"
	CodeDuplicateTask                    = "DUPLICATE_TASK"
	CodeInvalidCredentials               = "INVALID_CREDENTIALS"
	CodeInvalidCurrentPassword           = "INVALID_CURRENT_PASSWORD"
	CodeAccountDeactivated               = "ACCOUNT_DEACTIVATED"
	CodeTokenExpired                     = "TOKEN_EXPIRED"
	CodeTokenInvalid                     = "TOKEN_INVALID"
	CodeRefreshTokenInvalid              = "REFRESH_TOKEN_INVALID"
	CodeRefreshTokenRevoked              = "REFRESH_TOKEN_REVOKED"
	CodeRefreshTokenExpired              = "REFRESH_TOKEN_EXPIRED"
	CodeOAuthProviderInvalid             = "OAUTH_PROVIDER_INVALID"
	CodeOAuthIdentifierMissing           = "OAUTH_IDENTIFIER_MISSING"
	CodeOAuthAccountAlreadyLinked        = "OAUTH_ACCOUNT_ALREADY_LINKED"
	CodeOAuthLastAuthMethod              = "OAUTH_LAST_AUTH_METHOD"
	CodeOAuthEmailRequired               = "OAUTH_EMAIL_REQUIRED"
	CodeOAuthNotConfigured               = "OAUTH_NOT_CONFIGURED"
	CodeAppleTokenInvalid                = "APPLE_TOKEN_INVALID"
	CodeAppleTokenInvalidIssuer          = "APPLE_TOKEN_INVALID_ISSUER"
	CodeAppleTokenAudienceMismatch       = "APPLE_TOKEN_AUDIENCE_MISMATCH"
	CodeAccountDeleteInvalidConfirmation = "ACCOUNT_DELETE_INVALID_CONFIRMATION"
	CodeAvatarFileRequired               = "AVATAR_FILE_REQUIRED"
	CodeAvatarFileTooLarge               = "AVATAR_FILE_TOO_LARGE"
	CodeStorageNotConfigured             = "STORAGE_NOT_CONFIGURED"

	// Server error codes (5xx)
	CodeInternalError = "INTERNAL_SERVER_ERROR"
	CodeDatabaseError = "DATABASE_ERROR"
	CodeServiceError  = "SERVICE_ERROR"
)

// HTTP Status messages
const (
	MsgSuccess            = "Operation completed successfully"
	MsgCreated            = "Resource created successfully"
	MsgUpdated            = "Resource updated successfully"
	MsgDeleted            = "Resource deleted successfully"
	MsgBadRequest         = "Bad request"
	MsgValidationError    = "Validation error"
	MsgUnauthorized       = "Unauthorized access"
	MsgForbidden          = "Access forbidden"
	MsgNotFound           = "Resource not found"
	MsgConflict           = "Resource conflict"
	MsgInternalError      = "Internal server error"
	MsgDatabaseError      = "Database error occurred"
	MsgServiceUnavailable = "Service temporarily unavailable"
)
