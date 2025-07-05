// File: internal/utils/logger.go
// Brief: Custom logger handlers for slog
// Detailed: Contains CustomJSONHandler for formatting slog output with custom time formats
// Author: drama.lin@aver.com
// Date: 2025-07-05

package utils

import (
	"context"
	"io"
	"log/slog"
)

// CustomJSONHandler extends the standard JSONHandler with a custom time format
type CustomJSONHandler struct {
	*slog.JSONHandler
	timeFormat string
}

// NewCustomJSONHandler creates a new JSON handler with a custom time format
func NewCustomJSONHandler(w io.Writer, opts *slog.HandlerOptions) *CustomJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Create a copy of options to modify
	newOpts := *opts

	// Set a custom time formatter
	origReplaceAttr := opts.ReplaceAttr
	newOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		// Format the time attribute
		if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
			return slog.String(a.Key, a.Value.Time().Format("2006-01-02 15:04:05.000"))
		}

		// Call the original ReplaceAttr if it exists
		if origReplaceAttr != nil {
			return origReplaceAttr(groups, a)
		}

		return a
	}

	return &CustomJSONHandler{
		JSONHandler: slog.NewJSONHandler(w, &newOpts),
		timeFormat:  "2006-01-02 15:04:05.000", // Custom time format
	}
}

// Handle implements slog.Handler
func (h *CustomJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	// Use the underlying JSONHandler's Handle method
	return h.JSONHandler.Handle(ctx, r)
}

// WithAttrs implements slog.Handler
func (h *CustomJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomJSONHandler{
		JSONHandler: h.JSONHandler.WithAttrs(attrs).(*slog.JSONHandler),
		timeFormat:  h.timeFormat,
	}
}

// WithGroup implements slog.Handler
func (h *CustomJSONHandler) WithGroup(name string) slog.Handler {
	return &CustomJSONHandler{
		JSONHandler: h.JSONHandler.WithGroup(name).(*slog.JSONHandler),
		timeFormat:  h.timeFormat,
	}
}

// NewCustomTimeHandler creates a new JSON handler with the specified time format
func NewCustomTimeHandler(w io.Writer, opts *slog.HandlerOptions, timeFormat string) *CustomJSONHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	// Create a copy of options to modify
	newOpts := *opts

	// Set a custom time formatter
	origReplaceAttr := opts.ReplaceAttr
	newOpts.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
		// Format the time attribute
		if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
			return slog.String(a.Key, a.Value.Time().Format(timeFormat))
		}

		// Call the original ReplaceAttr if it exists
		if origReplaceAttr != nil {
			return origReplaceAttr(groups, a)
		}

		return a
	}

	return &CustomJSONHandler{
		JSONHandler: slog.NewJSONHandler(w, &newOpts),
		timeFormat:  timeFormat,
	}
}
