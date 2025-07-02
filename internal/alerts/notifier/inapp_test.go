package notifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInAppChannel(t *testing.T) {
	// Test with default size
	channel := NewInAppChannel(0)
	assert.NotNil(t, channel)
	assert.Equal(t, 100, channel.maxSize)

	// Test with custom size
	channel = NewInAppChannel(50)
	assert.NotNil(t, channel)
	assert.Equal(t, 50, channel.maxSize)
}

func TestInAppChannelSend(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)

	// Send a notification
	err := channel.Send(event, "Test Subject", "Test Body")
	require.NoError(t, err)

	// Verify the notification was stored
	notifications := channel.GetNotifications()
	require.Len(t, notifications, 1)
	assert.Equal(t, event.AlertID, notifications[0].AlertID)
	assert.Equal(t, event.Alert.Name, notifications[0].AlertName)
	assert.Equal(t, "Test Subject", notifications[0].Subject)
	assert.Equal(t, "Test Body", notifications[0].Message)
	assert.Equal(t, event.Alert.Severity, notifications[0].Severity)
	assert.Equal(t, event.NewState, notifications[0].State)
	assert.False(t, notifications[0].Read)
}

func TestInAppChannelMaxSize(t *testing.T) {
	channel := NewInAppChannel(3)
	event := createTestAlertEvent(t)

	// Send multiple notifications
	for i := 0; i < 5; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}

	// Verify only the most recent 3 are kept
	notifications := channel.GetNotifications()
	assert.Len(t, notifications, 3)
}

func TestInAppChannelGetUnreadNotifications(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)

	// Send multiple notifications
	for i := 0; i < 3; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}

	// Mark one as read
	notifications := channel.GetNotifications()
	channel.MarkAsRead(notifications[1].ID)

	// Get unread notifications
	unread := channel.GetUnreadNotifications()
	assert.Len(t, unread, 2)
}

func TestInAppChannelMarkAsRead(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)

	// Send a notification
	err := channel.Send(event, "Test Subject", "Test Body")
	require.NoError(t, err)

	// Get the notification ID
	notifications := channel.GetNotifications()
	require.Len(t, notifications, 1)
	id := notifications[0].ID

	// Mark as read
	result := channel.MarkAsRead(id)
	assert.True(t, result)

	// Verify it was marked as read
	notifications = channel.GetNotifications()
	assert.True(t, notifications[0].Read)

	// Try to mark a non-existent notification
	result = channel.MarkAsRead("non-existent-id")
	assert.False(t, result)
}

func TestInAppChannelMarkAllAsRead(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)

	// Send multiple notifications
	for i := 0; i < 3; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}

	// Mark all as read
	channel.MarkAllAsRead()

	// Verify all were marked as read
	notifications := channel.GetNotifications()
	for _, notification := range notifications {
		assert.True(t, notification.Read)
	}

	// Verify no unread notifications
	unread := channel.GetUnreadNotifications()
	assert.Len(t, unread, 0)
}

func TestInAppChannelClearNotifications(t *testing.T) {
	channel := NewInAppChannel(10)
	event := createTestAlertEvent(t)

	// Send multiple notifications
	for i := 0; i < 3; i++ {
		err := channel.Send(event, "Test Subject", "Test Body")
		require.NoError(t, err)
	}

	// Clear notifications
	channel.ClearNotifications()

	// Verify all notifications were removed
	notifications := channel.GetNotifications()
	assert.Len(t, notifications, 0)
}
