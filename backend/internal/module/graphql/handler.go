package graphqlmodule

import (
	"context"

	"golang/pkg/auth"
	authMiddleware "golang/pkg/middleware/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/graphql-go/graphql"
)

type Handler struct {
	schema     graphql.Schema
	jwtService *auth.JWTService
}

func NewHandler(schema graphql.Schema, jwtService *auth.JWTService) *Handler {
	return &Handler{schema: schema, jwtService: jwtService}
}

type postData struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// Handle handles both GET (GraphiQL playground in browser) and POST (JSON query execution).
func (h *Handler) Handle(c *fiber.Ctx) error {
	// If browser accesses GET /graphql, show interactive GraphiQL playground
	if c.Method() == fiber.MethodGet {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(graphiQLHTML)
	}

	var p postData
	if err := c.BodyParser(&p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors": []fiber.Map{{"message": "Invalid JSON payload"}},
		})
	}

	// Extract userID if available (from Bearer or OptionalAuth)
	userID, _ := authMiddleware.GetUserID(c)
	ctx := context.WithValue(c.UserContext(), "user_id", userID)

	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  p.Query,
		VariableValues: p.Variables,
		OperationName:  p.OperationName,
		Context:        ctx,
	})

	return c.JSON(result)
}

const graphiQLHTML = `<!DOCTYPE html>
<html>
<head>
  <title>TaskHub GraphQL Playground</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
  <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
  <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
  <div id="root"></div>
  <script>
    window.addEventListener('load', function (event) {
      GraphQLPlayground.init(document.getElementById('root'), {
        endpoint: '/api/v1/graphql',
        settings: {
          'request.credentials': 'include'
        }
      })
    })
  </script>
</body>
</html>`
