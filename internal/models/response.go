// File: internal/models/response.go
// Brief: API response models for Argus
// Detailed: Contains type definitions for APIResponse and Meta, used for consistent API responses.
// Author: Argus Migration (AI)
// Date: 2024-07-03

package models

// APIResponse represents a standard API response structure
// @brief Standard API response wrapper
// @author Argus
// @date 2024-07-03
type APIResponse struct {
	Success bool        `json:"success"`         // Indicates if the request was successful
	Data    interface{} `json:"data,omitempty"`  // Response data (optional)
	Error   string      `json:"error,omitempty"` // Error message (optional)
	Meta    *Meta       `json:"meta,omitempty"`  // Metadata (optional)
}

// Meta contains pagination and timing metadata for API responses
// @brief API response metadata
// @author Argus
// @date 2024-07-03
type Meta struct {
	Total     int    `json:"total"`     // Total number of items
	Page      int    `json:"page"`      // Current page number
	PerPage   int    `json:"per_page"`  // Items per page
	Timestamp string `json:"timestamp"` // Response timestamp (RFC3339)
}
