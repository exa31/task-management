package task

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTaskRequest_UnmarshalJSON(t *testing.T) {
	t.Run("null due_date sets ClearDueDate and HasDueDate", func(t *testing.T) {
		body := `{"due_date": null}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueDate)
		assert.True(t, req.ClearDueDate)
		assert.Nil(t, req.DueDate)
	})

	t.Run("empty string due_date sets ClearDueDate and HasDueDate", func(t *testing.T) {
		body := `{"due_date": ""}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueDate)
		assert.True(t, req.ClearDueDate)
		assert.Nil(t, req.DueDate)
	})

	t.Run("explicit clear_due_date flag sets ClearDueDate", func(t *testing.T) {
		body := `{"clear_due_date": true}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.ClearDueDate)
	})

	t.Run("valid due_date parses successfully", func(t *testing.T) {
		body := `{"due_date": "2026-10-25T15:00:00Z"}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueDate)
		assert.False(t, req.ClearDueDate)
		require.NotNil(t, req.DueDate)
		assert.Equal(t, 2026, req.DueDate.Year())
	})

	t.Run("date-only due_date parses successfully", func(t *testing.T) {
		body := `{"due_date": "2026-10-25"}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueDate)
		assert.False(t, req.ClearDueDate)
		require.NotNil(t, req.DueDate)
		assert.Equal(t, 2026, req.DueDate.Year())
		assert.Equal(t, time.Month(10), req.DueDate.Month())
		assert.Equal(t, 25, req.DueDate.Day())
	})

	t.Run("null project_id sets ClearProjectID and HasProjectID", func(t *testing.T) {
		body := `{"project_id": null}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasProjectID)
		assert.True(t, req.ClearProjectID)
		assert.Nil(t, req.ProjectID)
	})

	t.Run("zero project_id sets ClearProjectID and HasProjectID", func(t *testing.T) {
		body := `{"project_id": 0}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasProjectID)
		assert.True(t, req.ClearProjectID)
		assert.Nil(t, req.ProjectID)
	})

	t.Run("explicit clear_project_id flag sets ClearProjectID", func(t *testing.T) {
		body := `{"clear_project_id": true}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.ClearProjectID)
	})

	t.Run("empty string description sets ClearDescription", func(t *testing.T) {
		body := `{"description": ""}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.ClearDescription)
	})

	t.Run("null description sets ClearDescription", func(t *testing.T) {
		body := `{"description": null}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.ClearDescription)
	})

	t.Run("null recurrence_end_date sets ClearRecurrenceEndDate", func(t *testing.T) {
		body := `{"recurrence_end_date": null}`
		var req UpdateTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasRecurrenceEndDate)
		assert.True(t, req.ClearRecurrenceEndDate)
		assert.Nil(t, req.RecurrenceEndDate)
	})
}

func TestUpdateSubtaskRequest_UnmarshalJSON(t *testing.T) {
	t.Run("null due_date sets ClearDueDate and HasDueDate", func(t *testing.T) {
		body := `{"due_date": null}`
		var req UpdateSubtaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueDate)
		assert.True(t, req.ClearDueDate)
		assert.Nil(t, req.DueDate)
	})

	t.Run("empty string description sets ClearDescription", func(t *testing.T) {
		body := `{"description": ""}`
		var req UpdateSubtaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDescription)
		assert.True(t, req.ClearDescription)
	})
}

func TestUpdateRecurringTaskRequest_UnmarshalJSON(t *testing.T) {
	t.Run("null due_time sets ClearDueTime and HasDueTime", func(t *testing.T) {
		body := `{"due_time": null}`
		var req UpdateRecurringTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueTime)
		assert.True(t, req.ClearDueTime)
		assert.Nil(t, req.DueTime)
	})

	t.Run("empty string due_time sets ClearDueTime and HasDueTime", func(t *testing.T) {
		body := `{"due_time": ""}`
		var req UpdateRecurringTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasDueTime)
		assert.True(t, req.ClearDueTime)
		assert.Nil(t, req.DueTime)
	})

	t.Run("null project_id sets ClearProjectID and HasProjectID", func(t *testing.T) {
		body := `{"project_id": null}`
		var req UpdateRecurringTaskRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.True(t, req.HasProjectID)
		assert.True(t, req.ClearProjectID)
		assert.Nil(t, req.ProjectID)
	})
}
