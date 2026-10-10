package graphqlmodule

import (
	"context"
	"fmt"

	"golang/internal/module/note"
	"golang/internal/module/project"
	"golang/internal/module/task"
	usermodule "golang/internal/module/user"

	"github.com/graphql-go/graphql"
)

type SchemaConfig struct {
	UserRepo    usermodule.UserRepository
	ProjectRepo project.Repository
	TaskRepo    task.Repository
	NoteRepo    note.Repository
}

func BuildSchema(cfg SchemaConfig) (graphql.Schema, error) {
	// 1. User Type
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.String},
			"name":      &graphql.Field{Type: graphql.String},
			"email":     &graphql.Field{Type: graphql.String},
			"phone":     &graphql.Field{Type: graphql.String},
			"is_active": &graphql.Field{Type: graphql.Boolean},
		},
	})

	// 2. Project Type
	projectType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Project",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.Int},
			"name":        &graphql.Field{Type: graphql.String},
			"color":       &graphql.Field{Type: graphql.String},
			"description": &graphql.Field{Type: graphql.String},
			"task_count":  &graphql.Field{Type: graphql.Int},
			"created_at":  &graphql.Field{Type: graphql.String},
		},
	})

	// 3. Subtask Type
	subtaskType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Subtask",
		Fields: graphql.Fields{
			"id":           &graphql.Field{Type: graphql.Int},
			"title":        &graphql.Field{Type: graphql.String},
			"is_completed": &graphql.Field{Type: graphql.Boolean},
		},
	})

	// 4. Task Type
	taskType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Task",
		Fields: graphql.Fields{
			"id":                 &graphql.Field{Type: graphql.Int},
			"title":              &graphql.Field{Type: graphql.String},
			"description":        &graphql.Field{Type: graphql.String},
			"status":             &graphql.Field{Type: graphql.String},
			"priority":           &graphql.Field{Type: graphql.String},
			"is_focus":           &graphql.Field{Type: graphql.Boolean},
			"project_id":         &graphql.Field{Type: graphql.Int},
			"project_name":       &graphql.Field{Type: graphql.String},
			"project_color":      &graphql.Field{Type: graphql.String},
			"last_progress_note": &graphql.Field{Type: graphql.String},
			"subtasks_count":     &graphql.Field{Type: graphql.Int},
			"subtasks_completed": &graphql.Field{Type: graphql.Int},
			"progress_percent":   &graphql.Field{Type: graphql.Int},
			"subtasks":           &graphql.Field{Type: graphql.NewList(subtaskType)},
			"created_at":         &graphql.Field{Type: graphql.String},
		},
	})

	// 5. Note Type
	noteType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Note",
		Fields: graphql.Fields{
			"id":                &graphql.Field{Type: graphql.Int},
			"title":             &graphql.Field{Type: graphql.String},
			"content":           &graphql.Field{Type: graphql.String},
			"type":              &graphql.Field{Type: graphql.String},
			"color":             &graphql.Field{Type: graphql.String},
			"is_pinned":         &graphql.Field{Type: graphql.Boolean},
			"tags":              &graphql.Field{Type: graphql.NewList(graphql.String)},
			"converted_task_id": &graphql.Field{Type: graphql.Int},
			"created_at":        &graphql.Field{Type: graphql.String},
		},
	})

	// Root Query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"me": &graphql.Field{
				Type: userType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return nil, fmt.Errorf("unauthenticated")
					}
					u, err := cfg.UserRepo.GetByID(userID)
					if err != nil || u == nil {
						return nil, err
					}
					return map[string]interface{}{
						"id":        u.ID,
						"name":      u.Name,
						"email":     u.Email,
						"phone":     u.Phone.String,
						"is_active": u.IsActive,
					}, nil
				},
			},
			"projects": &graphql.Field{
				Type: graphql.NewList(projectType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return []interface{}{}, nil
					}
					return cfg.ProjectRepo.FindAll(context.Background(), userID)
				},
			},
			"tasks": &graphql.Field{
				Type: graphql.NewList(taskType),
				Args: graphql.FieldConfigArgument{
					"projectId": &graphql.ArgumentConfig{Type: graphql.Int},
					"status":    &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return []interface{}{}, nil
					}
					filter := task.TaskFilter{}
					if status, ok := p.Args["status"].(string); ok {
						filter.Status = status
					}
					if pID, ok := p.Args["projectId"].(int); ok {
						val := int64(pID)
						filter.ProjectID = &val
					}
					return cfg.TaskRepo.FindAll(context.Background(), userID, filter)
				},
			},
			"notes": &graphql.Field{
				Type: graphql.NewList(noteType),
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return []interface{}{}, nil
					}
					filter := note.NoteFilter{}
					if nType, ok := p.Args["type"].(string); ok {
						filter.Type = nType
					}
					return cfg.NoteRepo.FindAll(context.Background(), userID, filter)
				},
			},
		},
	})

	// Root Mutation
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createProject": &graphql.Field{
				Type: projectType,
				Args: graphql.FieldConfigArgument{
					"name":        &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"color":       &graphql.ArgumentConfig{Type: graphql.String},
					"description": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return nil, fmt.Errorf("unauthenticated")
					}
					name := p.Args["name"].(string)
					color, _ := p.Args["color"].(string)
					desc, _ := p.Args["description"].(string)
					return cfg.ProjectRepo.Create(context.Background(), userID, &project.CreateProjectRequest{
						Name:        name,
						Color:       color,
						Description: desc,
					})
				},
			},
			"createTask": &graphql.Field{
				Type: taskType,
				Args: graphql.FieldConfigArgument{
					"title":       &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"description": &graphql.ArgumentConfig{Type: graphql.String},
					"priority":    &graphql.ArgumentConfig{Type: graphql.String},
					"projectId":   &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return nil, fmt.Errorf("unauthenticated")
					}
					title := p.Args["title"].(string)
					desc, _ := p.Args["description"].(string)
					priority, _ := p.Args["priority"].(string)
					var pID *int64
					if v, ok := p.Args["projectId"].(int); ok {
						val := int64(v)
						pID = &val
					}
					return cfg.TaskRepo.Create(context.Background(), userID, task.CreateTaskRequest{
						Title:       title,
						Description: desc,
						Priority:    priority,
						ProjectID:   pID,
					})
				},
			},
			"createNote": &graphql.Field{
				Type: noteType,
				Args: graphql.FieldConfigArgument{
					"title":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"content": &graphql.ArgumentConfig{Type: graphql.String},
					"type":    &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return nil, fmt.Errorf("unauthenticated")
					}
					title := p.Args["title"].(string)
					content, _ := p.Args["content"].(string)
					nType, _ := p.Args["type"].(string)
					if nType == "" {
						nType = "actionable"
					}
					return cfg.NoteRepo.Create(context.Background(), userID, note.CreateNoteRequest{
						Title:   title,
						Content: content,
						Type:    nType,
					})
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}
