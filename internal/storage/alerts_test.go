package storage

import (
	"os"
	"testing"
	"time"

	"argus/internal/database"
	"argus/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAlertStore(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test with custom config directory
	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create and retrieve an alert to verify store is operational
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	retrieved, err := store.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.Equal(t, alert.ID, retrieved.ID)

	// Test with default config directory
	defaultStore, err := database.NewAlertStore("")
	require.NoError(t, err)
	defaultAlert := createTestAlert()
	err = defaultStore.CreateAlert(defaultAlert)
	require.NoError(t, err)

	defaultRetrieved, err := defaultStore.GetAlert(defaultAlert.ID)
	require.NoError(t, err)
	assert.Equal(t, defaultAlert.ID, defaultRetrieved.ID)

	defer os.RemoveAll(database.DefaultConfigDir)
}

func TestAlertStore_CreateAlert(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create a valid alert
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	// Verify the alert was created and can be retrieved
	retrieved, err := store.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.Equal(t, alert.ID, retrieved.ID)
	assert.Equal(t, alert.Name, retrieved.Name)
	assert.Equal(t, alert.Description, retrieved.Description)
	assert.Equal(t, alert.Severity, retrieved.Severity)

	// Try to create the same alert again (should fail)
	err = store.CreateAlert(alert)
	assert.Error(t, err)

	// Create an alert with an invalid configuration
	invalidAlert := createTestAlert()
	invalidAlert.ID = ""
	invalidAlert.Name = ""
	err = store.CreateAlert(invalidAlert)
	assert.Error(t, err)
	assert.NotEmpty(t, invalidAlert.ID, "ID should be generated for empty ID")
}

func TestAlertStore_GetAlert(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create a test alert
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	// Get the alert
	retrieved, err := store.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.Equal(t, alert.ID, retrieved.ID)
	assert.Equal(t, alert.Name, retrieved.Name)
	assert.Equal(t, alert.Description, retrieved.Description)
	assert.Equal(t, alert.Severity, retrieved.Severity)

	// Try to get a non-existent alert
	_, err = store.GetAlert("non-existent-id")
	assert.ErrorIs(t, err, database.ErrAlertNotFound)

	// Try to get with an empty ID
	_, err = store.GetAlert("")
	assert.ErrorIs(t, err, database.ErrInvalidAlertID)
}

func TestAlertStore_UpdateAlert(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create a test alert
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	// Update the alert
	alert.Name = "Updated Alert Name"
	alert.Description = "Updated description"
	alert.Severity = models.SeverityCritical
	err = store.UpdateAlert(alert)
	require.NoError(t, err)

	// Get the updated alert
	retrieved, err := store.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Alert Name", retrieved.Name)
	assert.Equal(t, "Updated description", retrieved.Description)
	assert.Equal(t, models.SeverityCritical, retrieved.Severity)

	// Verify backup was created
	backupFiles, err := store.ListBackups(alert.ID)
	require.NoError(t, err)
	assert.Len(t, backupFiles, 1)

	// Try to update a non-existent alert
	nonExistentAlert := createTestAlert()
	nonExistentAlert.ID = "non-existent-id"
	err = store.UpdateAlert(nonExistentAlert)
	assert.ErrorIs(t, err, database.ErrAlertNotFound)

	// Try to update with an empty ID
	invalidAlert := createTestAlert()
	invalidAlert.ID = ""
	err = store.UpdateAlert(invalidAlert)
	assert.ErrorIs(t, err, database.ErrInvalidAlertID)
}

func TestAlertStore_DeleteAlert(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create a test alert
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	// Delete the alert
	err = store.DeleteAlert(alert.ID)
	require.NoError(t, err)

	// Verify the alert was deleted
	_, err = store.GetAlert(alert.ID)
	assert.ErrorIs(t, err, database.ErrAlertNotFound)

	// Verify backup was created
	backupFiles, err := store.ListBackups(alert.ID)
	require.NoError(t, err)
	assert.Len(t, backupFiles, 1)

	// Try to delete a non-existent alert
	err = store.DeleteAlert("non-existent-id")
	assert.ErrorIs(t, err, database.ErrAlertNotFound)

	// Try to delete with an empty ID
	err = store.DeleteAlert("")
	assert.ErrorIs(t, err, database.ErrInvalidAlertID)
}

func TestAlertStore_ListAlerts(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create multiple test alerts
	alert1 := createTestAlert()
	alert1.Name = "Alert 1"
	err = store.CreateAlert(alert1)
	require.NoError(t, err)

	alert2 := createTestAlert()
	alert2.Name = "Alert 2"
	err = store.CreateAlert(alert2)
	require.NoError(t, err)

	alert3 := createTestAlert()
	alert3.Name = "Alert 3"
	err = store.CreateAlert(alert3)
	require.NoError(t, err)

	// List all alerts
	alerts, err := store.ListAlerts()
	require.NoError(t, err)
	assert.Len(t, alerts, 3)

	// Verify alert IDs are in the list
	alertIDs := make(map[string]bool)
	for _, a := range alerts {
		alertIDs[a.ID] = true
	}
	assert.True(t, alertIDs[alert1.ID])
	assert.True(t, alertIDs[alert2.ID])
	assert.True(t, alertIDs[alert3.ID])
}

func TestAlertStore_RestoreAlert(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create a test alert
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	// Update the alert to create a backup
	alert.Name = "Updated Alert Name"
	err = store.UpdateAlert(alert)
	require.NoError(t, err)

	// Get the backup timestamp
	backups, err := store.ListBackups(alert.ID)
	require.NoError(t, err)
	require.Len(t, backups, 1)
	timestamp := backups[0]

	// Restore the alert from backup
	err = store.RestoreAlert(alert.ID, timestamp)
	require.NoError(t, err)

	// Get the restored alert
	restored, err := store.GetAlert(alert.ID)
	require.NoError(t, err)
	assert.NotEqual(t, "Updated Alert Name", restored.Name)
	assert.Equal(t, alert.ID, restored.ID)

	// Try to restore a non-existent backup
	err = store.RestoreAlert(alert.ID, "non-existent-timestamp")
	assert.Error(t, err)

	// Try to restore with an empty ID
	err = store.RestoreAlert("", timestamp)
	assert.ErrorIs(t, err, database.ErrInvalidAlertID)
}

func TestAlertStore_ListBackups(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "alertstore_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	store, err := database.NewAlertStore(tempDir)
	require.NoError(t, err)

	// Create a test alert
	alert := createTestAlert()
	err = store.CreateAlert(alert)
	require.NoError(t, err)

	// Create multiple backups by updating the alert
	for i := 0; i < 3; i++ {
		alert.Name = "Updated Alert Name " + time.Now().String()
		err = store.UpdateAlert(alert)
		require.NoError(t, err)
		time.Sleep(time.Second) // Wait 1 second between updates to ensure unique backup timestamps
	}

	// List backups
	backups, err := store.ListBackups(alert.ID)
	require.NoError(t, err)
	assert.Len(t, backups, 3)

	// Try to list backups for a non-existent alert ID
	backups, err = store.ListBackups("non-existent-id")
	require.NoError(t, err)
	assert.Empty(t, backups)

	// Try to list backups with an empty ID
	_, err = store.ListBackups("")
	assert.ErrorIs(t, err, database.ErrInvalidAlertID)
}

// Helper functions

func createTestAlert() *models.AlertConfig {
	now := time.Now()
	return &models.AlertConfig{
		ID:          uuid.New().String(),
		Name:        "Test Alert",
		Description: "Test alert description",
		Enabled:     true,
		Severity:    models.SeverityWarning,
		Threshold: models.ThresholdConfig{
			MetricType: models.MetricCPU,
			MetricName: "usage_percent",
			Operator:   models.OperatorGreaterThan,
			Value:      90.0,
			Duration:   5 * time.Minute,
		},
		Notifications: []models.NotificationConfig{
			{
				Type:    models.NotificationInApp,
				Enabled: true,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
