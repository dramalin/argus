package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchedule_Validate(t *testing.T) {
	s := &Schedule{CronExpression: "* * * * *"}
	assert.NoError(t, s.Validate())

	s = &Schedule{OneTime: true}
	assert.NoError(t, s.Validate())

	s = &Schedule{}
	assert.Error(t, s.Validate())
}

func TestTaskConfig_Validate(t *testing.T) {
	task := &TaskConfig{
		ID:        GenerateID(),
		Name:      "Test",
		Type:      TaskLogRotation,
		Schedule:  Schedule{CronExpression: "* * * * *"},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	assert.NoError(t, task.Validate())

	task.Type = "invalid_type"
	assert.Error(t, task.Validate())

	task.Type = TaskLogRotation
	task.ID = ""
	assert.Error(t, task.Validate())
}
