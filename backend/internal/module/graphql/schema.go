package graphqlmodule

import (
	"context"
	"fmt"
	"time"

	"golang/database"
	"golang/internal/module/note"
	"golang/internal/module/project"
	"golang/internal/module/reminder"
	"golang/internal/module/task"
	usermodule "golang/internal/module/user"

	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

type SchemaConfig struct {
	UserRepo     usermodule.UserRepository
	ProjectRepo  project.Repository
	TaskRepo     task.Repository
	NoteRepo     note.Repository
	ReminderRepo reminder.Repository
	DB           database.Transactor
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
			"description":  &graphql.Field{Type: graphql.String},
			"is_completed": &graphql.Field{Type: graphql.Boolean},
			"due_date": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if st, ok := p.Source.(task.SubtaskResponse); ok && st.DueDate != nil {
						return st.DueDate.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
		},
	})

	// 4. Task Type
	taskType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Task",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.Int},
			"title":       &graphql.Field{Type: graphql.String},
			"description": &graphql.Field{Type: graphql.String},
			"status":      &graphql.Field{Type: graphql.String},
			"priority":    &graphql.Field{Type: graphql.String},
			"due_date": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if t, ok := p.Source.(task.TaskResponse); ok && t.DueDate != nil {
						return t.DueDate.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"is_focus":      &graphql.Field{Type: graphql.Boolean},
			"project_id":    &graphql.Field{Type: graphql.Int},
			"project_name":  &graphql.Field{Type: graphql.String},
			"project_color": &graphql.Field{Type: graphql.String},
			"project": &graphql.Field{
				Type: projectType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if t, ok := p.Source.(task.TaskResponse); ok && t.ProjectID != nil {
						name := ""
						if t.ProjectName != nil {
							name = *t.ProjectName
						}
						color := ""
						if t.ProjectColor != nil {
							color = *t.ProjectColor
						}
						return map[string]interface{}{
							"id":    *t.ProjectID,
							"name":  name,
							"color": color,
						}, nil
					}
					return nil, nil
				},
			},
			"tags": &graphql.Field{
				Type: graphql.NewList(graphql.String),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if t, ok := p.Source.(task.TaskResponse); ok {
						if t.Tags == nil {
							return []string{}, nil
						}
						return t.Tags, nil
					}
					return []string{}, nil
				},
			},
			"last_progress_note": &graphql.Field{Type: graphql.String},
			"subtasks_count":     &graphql.Field{Type: graphql.Int},
			"subtasks_completed": &graphql.Field{Type: graphql.Int},
			"progress_percent":   &graphql.Field{Type: graphql.Int},
			"progress": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if t, ok := p.Source.(task.TaskResponse); ok {
						return t.ProgressPercent, nil
					}
					return 0, nil
				},
			},
			"subtasks":   &graphql.Field{Type: graphql.NewList(subtaskType)},
			"created_at": &graphql.Field{Type: graphql.String},
		},
	})

	// 5. Reminder Type
	reminderType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Reminder",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.Int},
			"title":       &graphql.Field{Type: graphql.String},
			"description": &graphql.Field{Type: graphql.String},
			"remind_at": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if r, ok := p.Source.(reminder.ReminderResponse); ok {
						return r.RemindAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"status":       &graphql.Field{Type: graphql.String},
			"task_id":      &graphql.Field{Type: graphql.Int},
			"task_title":   &graphql.Field{Type: graphql.String},
			"snooze_count": &graphql.Field{Type: graphql.Int},
			"created_at":   &graphql.Field{Type: graphql.String},
		},
	})

	// 6. Task Activity Type
	activityType := graphql.NewObject(graphql.ObjectConfig{
		Name: "TaskActivity",
		Fields: graphql.Fields{
			"id":            &graphql.Field{Type: graphql.Int},
			"task_id":       &graphql.Field{Type: graphql.Int},
			"task_title":    &graphql.Field{Type: graphql.String},
			"activity_type": &graphql.Field{Type: graphql.String},
			"description":   &graphql.Field{Type: graphql.String},
			"created_at": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if a, ok := p.Source.(task.TaskActivityResponse); ok {
						return a.CreatedAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
		},
	})

	// 7. Note Type
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

	// 8. Date Count Type
	dateCountType := graphql.NewObject(graphql.ObjectConfig{
		Name: "DateCount",
		Fields: graphql.Fields{
			"date":  &graphql.Field{Type: graphql.String},
			"count": &graphql.Field{Type: graphql.Int},
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
					return cfg.ProjectRepo.FindAll(p.Context, nil, userID)
				},
			},
			"tasks": &graphql.Field{
				Type: graphql.NewList(taskType),
				Args: graphql.FieldConfigArgument{
					"projectId": &graphql.ArgumentConfig{Type: graphql.Int},
					"status":    &graphql.ArgumentConfig{Type: graphql.String},
					"from":      &graphql.ArgumentConfig{Type: graphql.String},
					"to":        &graphql.ArgumentConfig{Type: graphql.String},
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
					if fStr, ok := p.Args["from"].(string); ok && fStr != "" {
						if t, err := time.Parse(time.RFC3339, fStr); err == nil {
							filter.From = &t
						}
					}
					if tStr, ok := p.Args["to"].(string); ok && tStr != "" {
						if t, err := time.Parse(time.RFC3339, tStr); err == nil {
							filter.To = &t
						}
					}
					var items []task.TaskResponse
					if cfg.DB != nil {
						_ = cfg.DB.WithReadOnlyTransactionCtx(p.Context, func(tx *sqlx.Tx) error {
							var err error
							items, err = cfg.TaskRepo.FindAll(p.Context, tx, userID, filter)
							return err
						})
					}
					return items, nil
				},
			},
			"reminders": &graphql.Field{
				Type: graphql.NewList(reminderType),
				Args: graphql.FieldConfigArgument{
					"status": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if cfg.ReminderRepo == nil {
						return []interface{}{}, nil
					}
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return []interface{}{}, nil
					}
					filter := reminder.ReminderFilter{}
					if status, ok := p.Args["status"].(string); ok && status != "" {
						filter.Status = status
					}
					items, _, err := cfg.ReminderRepo.FindAll(context.Background(), userID, filter)
					if err != nil {
						return []interface{}{}, nil
					}
					return items, nil
				},
			},
			"activities": &graphql.Field{
				Type: graphql.NewList(activityType),
				Args: graphql.FieldConfigArgument{
					"taskId": &graphql.ArgumentConfig{Type: graphql.Int},
					"from":   &graphql.ArgumentConfig{Type: graphql.String},
					"to":     &graphql.ArgumentConfig{Type: graphql.String},
					"date":   &graphql.ArgumentConfig{Type: graphql.String},
					"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return []interface{}{}, nil
					}
					if tID, ok := p.Args["taskId"].(int); ok && tID > 0 {
						var acts []task.TaskActivityResponse
						if cfg.DB != nil {
							_ = cfg.DB.WithReadOnlyTransactionCtx(p.Context, func(tx *sqlx.Tx) error {
								var err error
								acts, err = cfg.TaskRepo.FindActivitiesByTaskID(p.Context, tx, userID, int64(tID))
								return err
							})
						}
						return acts, nil
					}
					var fromTime, toTime *time.Time
					if fStr, ok := p.Args["from"].(string); ok && fStr != "" {
						if t, err := time.Parse(time.RFC3339, fStr); err == nil {
							fromTime = &t
						}
					}
					if tStr, ok := p.Args["to"].(string); ok && tStr != "" {
						if t, err := time.Parse(time.RFC3339, tStr); err == nil {
							toTime = &t
						}
					}
					if dateStr, ok := p.Args["date"].(string); ok && dateStr != "" {
						if dt, err := time.Parse("2006-01-02", dateStr); err == nil {
							loc := time.FixedZone("WIB", 7*3600)
							startOfDay := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, loc).UTC()
							endOfDay := time.Date(dt.Year(), dt.Month(), dt.Day(), 23, 59, 59, 999999999, loc).UTC()
							fromTime = &startOfDay
							toTime = &endOfDay
						}
					}
					var limitArgs []int
					if l, ok := p.Args["limit"].(int); ok && l > 0 {
						limitArgs = append(limitArgs, l)
					}
					var acts []task.TaskActivityResponse
					if cfg.DB != nil {
						_ = cfg.DB.WithReadOnlyTransactionCtx(p.Context, func(tx *sqlx.Tx) error {
							var err error
							acts, err = cfg.TaskRepo.FindActivitiesByUser(p.Context, tx, userID, fromTime, toTime, limitArgs...)
							return err
						})
					}
					return acts, nil
				},
			},
			"activityCounts": &graphql.Field{
				Type: graphql.NewList(dateCountType),
				Args: graphql.FieldConfigArgument{
					"from":     &graphql.ArgumentConfig{Type: graphql.String},
					"to":       &graphql.ArgumentConfig{Type: graphql.String},
					"timezone": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userID, ok := p.Context.Value("user_id").(string)
					if !ok || userID == "" {
						return []interface{}{}, nil
					}
					var fromTime, toTime *time.Time
					if fStr, ok := p.Args["from"].(string); ok && fStr != "" {
						if t, err := time.Parse(time.RFC3339, fStr); err == nil {
							fromTime = &t
						}
					}
					if tStr, ok := p.Args["to"].(string); ok && tStr != "" {
						if t, err := time.Parse(time.RFC3339, tStr); err == nil {
							toTime = &t
						}
					}
					tz := "Asia/Jakarta"
					if tStr, ok := p.Args["timezone"].(string); ok && tStr != "" {
						tz = tStr
					}
					var counts []task.ActivityDateCount
					if cfg.DB != nil {
						_ = cfg.DB.WithReadOnlyTransactionCtx(p.Context, func(tx *sqlx.Tx) error {
							var err error
							counts, err = cfg.TaskRepo.FindActivityCountsByUser(p.Context, tx, userID, fromTime, toTime, tz)
							return err
						})
					}
					return counts, nil
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
					if cfg.DB != nil {
						var created *project.ProjectResponse
						err := cfg.DB.WithTransactionCtx(p.Context, func(tx *sqlx.Tx) error {
							var err error
							created, err = cfg.ProjectRepo.Create(p.Context, tx, userID, &project.CreateProjectRequest{
								Name:        name,
								Color:       color,
								Description: desc,
							})
							return err
						})
						if err != nil {
							return nil, err
						}
						return created, nil
					}
					return cfg.ProjectRepo.Create(p.Context, nil, userID, &project.CreateProjectRequest{
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
					var created *task.TaskResponse
					if cfg.DB != nil {
						err := cfg.DB.WithTransactionCtx(p.Context, func(tx *sqlx.Tx) error {
							txCtx := database.ContextWithTx(p.Context, tx)
							var err error
							created, err = cfg.TaskRepo.Create(txCtx, tx, userID, task.CreateTaskRequest{
								Title:       title,
								Description: desc,
								Priority:    priority,
								ProjectID:   pID,
							})
							return err
						})
						if err != nil {
							return nil, err
						}
					}
					return created, nil
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
					return cfg.NoteRepo.Create(p.Context, userID, note.CreateNoteRequest{
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
