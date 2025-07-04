package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   APIResponse
		wantErr bool
	}{
		{
			name: "Valid response with data",
			input: APIResponse{
				Success: true,
				Data:    map[string]string{"key": "value"},
				Meta: &Meta{
					Total:     100,
					Page:      1,
					PerPage:   10,
					Timestamp: "2024-07-03T12:00:00Z",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid response with error",
			input: APIResponse{
				Success: false,
				Error:   "Something went wrong",
				Meta: &Meta{
					Total:     0,
					Page:      1,
					PerPage:   10,
					Timestamp: "2024-07-03T12:00:00Z",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Test JSON unmarshaling
			var got APIResponse
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			// Compare the original and unmarshaled data
			assert.Equal(t, tt.input.Success, got.Success)
			if tt.input.Data != nil {
				// JSON unmarshaling converts map[string]string to map[string]interface{}
				// Convert both to JSON and compare to ensure equal values
				inputJSON, err := json.Marshal(tt.input.Data)
				require.NoError(t, err)
				gotJSON, err := json.Marshal(got.Data)
				require.NoError(t, err)
				assert.JSONEq(t, string(inputJSON), string(gotJSON))
			}
			if tt.input.Error != "" {
				assert.Equal(t, tt.input.Error, got.Error)
			}
			if tt.input.Meta != nil {
				assert.Equal(t, tt.input.Meta.Total, got.Meta.Total)
				assert.Equal(t, tt.input.Meta.Page, got.Meta.Page)
				assert.Equal(t, tt.input.Meta.PerPage, got.Meta.PerPage)
				assert.Equal(t, tt.input.Meta.Timestamp, got.Meta.Timestamp)
			}
		})
	}
}

func TestMeta(t *testing.T) {
	tests := []struct {
		name    string
		input   Meta
		wantErr bool
	}{
		{
			name: "Valid metadata",
			input: Meta{
				Total:     100,
				Page:      1,
				PerPage:   10,
				Timestamp: "2024-07-03T12:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "Zero total count",
			input: Meta{
				Total:     0,
				Page:      1,
				PerPage:   10,
				Timestamp: "2024-07-03T12:00:00Z",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			data, err := json.Marshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Test JSON unmarshaling
			var got Meta
			err = json.Unmarshal(data, &got)
			require.NoError(t, err)

			// Compare the original and unmarshaled data
			assert.Equal(t, tt.input.Total, got.Total)
			assert.Equal(t, tt.input.Page, got.Page)
			assert.Equal(t, tt.input.PerPage, got.PerPage)
			assert.Equal(t, tt.input.Timestamp, got.Timestamp)
		})
	}
}
