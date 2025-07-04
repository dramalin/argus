package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"argus/internal/config"
)

// MockRoutesRegister is a mock for the IRoutesRegister interface
type MockRoutesRegister struct {
	mock.Mock
}

func (m *MockRoutesRegister) RegisterRoutes(router *gin.RouterGroup) {
	m.Called(router)
}

func TestNewServer(t *testing.T) {
	// Setup mocks
	mockCfg := &config.Config{}
	mockAlertHandler := new(MockRoutesRegister)
	mockTaskHandler := new(MockRoutesRegister)

	// Create mock handler functions
	mockCPUHandler := func(c *gin.Context) {}
	mockMemoryHandler := func(c *gin.Context) {}
	mockNetworkHandler := func(c *gin.Context) {}
	mockProcessHandler := func(c *gin.Context) {}

	// Set up expectations
	mockAlertHandler.On("RegisterRoutes", mock.Anything).Return()
	mockTaskHandler.On("RegisterRoutes", mock.Anything).Return()

	// Create a new server
	server := NewServer(mockCfg, mockAlertHandler, mockTaskHandler,
		mockCPUHandler, mockMemoryHandler, mockNetworkHandler, mockProcessHandler)

	// Assert server is not nil
	assert.NotNil(t, server)

	// Verify that the routes have been registered
	routes := server.Routes()
	assert.NotEmpty(t, routes)

	mockAlertHandler.AssertExpectations(t)
	mockTaskHandler.AssertExpectations(t)
}
