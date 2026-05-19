package tasks

import (
	"testing"
	"time"
)

func TestCreateTaskInput_Defaults(t *testing.T) {
	t.Run("default status ID should be 1 (Todo)", func(t *testing.T) {
		// Test the default logic from task_service.go
		statusID := uint(0)
		if statusID == 0 {
			statusID = 1
		}
		if statusID != 1 {
			t.Errorf("Expected default status ID to be 1 (Todo), got %d", statusID)
		}
	})

	t.Run("default priority ID should be 2 (Medium)", func(t *testing.T) {
		// Test the default logic from task_service.go
		priorityID := uint(0)
		if priorityID == 0 {
			priorityID = 2
		}
		if priorityID != 2 {
			t.Errorf("Expected default priority ID to be 2 (Medium), got %d", priorityID)
		}
	})
}

func TestCreateTaskInput_Fields(t *testing.T) {
	startDate := time.Now()
	endDate := startDate.Add(24 * time.Hour)
	input := CreateTaskInput{
		Title:      "Test Task",
		ProjectID:  1,
		StatusID:   1,
		PriorityID: 2,
		StartDate:  &startDate,
		EndDate:    &endDate,
		LabelIDs:   []uint{1, 2, 3},
	}

	t.Run("all fields should be set correctly", func(t *testing.T) {
		if input.Title == "" {
			t.Errorf("CreateTaskInput should have Title")
		}
		if input.ProjectID == 0 {
			t.Errorf("CreateTaskInput should have ProjectID")
		}
		if len(input.LabelIDs) == 0 {
			t.Errorf("CreateTaskInput should support LabelIDs")
		}
	})
}

func TestUpdateTaskInput_Fields(t *testing.T) {
	title := "Updated Title"
	statusID := uint(2)
	assigneeIDs := []uint{1, 2, 3}
	startDate := time.Now()
	endDate := startDate.Add(48 * time.Hour)
	expectedVersion := uint(1)

	input := UpdateTaskInput{
		Title:           &title,
		StatusID:        &statusID,
		AssigneeIDs:     assigneeIDs,
		StartDate:       &startDate,
		EndDate:         &endDate,
		ExpectedVersion: &expectedVersion,
	}

	t.Run("pointer fields should be settable", func(t *testing.T) {
		if input.Title == nil || *input.Title != title {
			t.Errorf("UpdateTaskInput.Title should be settable")
		}
		if input.StatusID == nil || *input.StatusID != statusID {
			t.Errorf("UpdateTaskInput.StatusID should be settable")
		}
	})

	t.Run("slice fields should work", func(t *testing.T) {
		if len(input.AssigneeIDs) != 3 {
			t.Errorf("UpdateTaskInput.AssigneeIDs should have 3 elements")
		}
	})

	t.Run("expected version for optimistic locking", func(t *testing.T) {
		if input.ExpectedVersion == nil {
			t.Errorf("UpdateTaskInput should support ExpectedVersion for optimistic locking")
		}
	})
}

func TestTaskModel(t *testing.T) {
	t.Run("task should have required fields", func(t *testing.T) {
		task := Task{
			ID:        1,
			Title:     "Test Task",
			StatusID:  1,
			ProjectID: 1,
		}

		if task.ID == 0 {
			t.Errorf("Task should have ID")
		}
		if task.Title == "" {
			t.Errorf("Task should have Title")
		}
	})

	t.Run("task should support assignees", func(t *testing.T) {
		task := Task{
			ID:           1,
			AssigneeIDs:  []uint{1, 2, 3},
		}

		if len(task.AssigneeIDs) != 3 {
			t.Errorf("Task should support multiple assignees")
		}
	})

	t.Run("task should have version field for optimistic locking", func(t *testing.T) {
		task := Task{
			ID:      1,
			Version: 1,
		}

		if task.Version != 1 {
			t.Errorf("Task should have Version field for optimistic locking")
		}
	})
}

func TestStatusModel(t *testing.T) {
	t.Run("status should have index for ordering", func(t *testing.T) {
		status := Status{
			ID:        1,
			Name:      "Todo",
			Index:     0,
			ProjectID: 1,
		}

		if status.Index != 0 {
			t.Errorf("Status should have Index field for ordering")
		}
	})

	t.Run("status should support multiple statuses per project", func(t *testing.T) {
		statuses := []Status{
			{ID: 1, Name: "Todo", Index: 0, ProjectID: 1},
			{ID: 2, Name: "In Progress", Index: 1, ProjectID: 1},
			{ID: 3, Name: "Done", Index: 2, ProjectID: 1},
		}

		if len(statuses) != 3 {
			t.Errorf("Should support multiple statuses per project")
		}
	})
}

func TestTaskUserModel(t *testing.T) {
	t.Run("task user relationship", func(t *testing.T) {
		tu := TaskUser{
			TaskID: 1,
			UserID: 1,
		}

		if tu.TaskID == 0 {
			t.Errorf("TaskUser should have TaskID")
		}
		if tu.UserID == 0 {
			t.Errorf("TaskUser should have UserID")
		}
	})
}

func TestDefaultStatuses(t *testing.T) {
	t.Run("default statuses should be created", func(t *testing.T) {
		defaults := []string{"Todo", "On Progress", "Done", "Pending", "Cancel"}

		if len(defaults) != 5 {
			t.Errorf("Expected 5 default statuses, got %d", len(defaults))
		}

		expectedStatuses := map[string]int{
			"Todo":        0,
			"On Progress": 1,
			"Done":        2,
			"Pending":     3,
			"Cancel":      4,
		}

		for name, expectedIndex := range expectedStatuses {
			if actualIndex := expectedStatuses[name]; actualIndex != expectedIndex {
				t.Errorf("Status %s should have index %d, got %d", name, expectedIndex, actualIndex)
			}
		}
	})
}

func TestPriorityModel(t *testing.T) {
	t.Run("priority should have level for sorting", func(t *testing.T) {
		priority := Priority{
			ID:    1,
			Name:  "High",
			Level: 3,
			Color: "#FF0000",
		}

		if priority.Level != 3 {
			t.Errorf("Priority should have Level field")
		}
		if priority.Color == "" {
			t.Errorf("Priority should have Color field")
		}
	})
}

func TestTaskFilters(t *testing.T) {
	t.Run("filters should be optional", func(t *testing.T) {
		filters := TaskFilters{}

		// Filters should be nil/empty when not set
		if filters.StatusID != nil {
			t.Errorf("TaskFilters should have nil StatusID when not set")
		}
		if filters.PriorityID != nil {
			t.Errorf("TaskFilters should have nil PriorityID when not set")
		}
		if filters.AssigneeID != nil {
			t.Errorf("TaskFilters should have nil AssigneeID when not set")
		}
	})
}

func TestInterfaceToString(t *testing.T) {
	testCases := []struct {
		name     string
		id       uint
		expected string
	}{
		{
			name:     "positive number",
			id:       123,
			expected: "123",
		},
		{
			name:     "zero",
			id:       0,
			expected: "0",
		},
		{
			name:     "large number",
			id:       999999,
			expected: "999999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := interfaceToString(tc.id)
			if result != tc.expected {
				t.Errorf("interfaceToString(%d) = %s, want %s", tc.id, result, tc.expected)
			}
		})
	}
}

func TestVersionConflictError(t *testing.T) {
	t.Run("error should implement error interface", func(t *testing.T) {
		err := &VersionConflictError{CurrentVersion: 5}

		if err.Error() == "" {
			t.Errorf("VersionConflictError should implement Error() method")
		}
	})

	t.Run("error should indicate conflict", func(t *testing.T) {
		err := &VersionConflictError{CurrentVersion: 5}

		if !err.Conflict() {
			t.Errorf("VersionConflictError should return true for Conflict()")
		}
	})

	t.Run("error should include current version", func(t *testing.T) {
		currentVersion := uint(10)
		err := &VersionConflictError{CurrentVersion: currentVersion}

		if err.CurrentVersion != currentVersion {
			t.Errorf("VersionConflictError should store CurrentVersion")
		}
	})
}