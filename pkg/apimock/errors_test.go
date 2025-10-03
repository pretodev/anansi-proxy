package apimock

import (
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "Full error with filename and line",
			err: &ParseError{
				Filename: "test.apimock",
				Line:     42,
				Message:  "syntax error",
			},
			expected: "test.apimock:42: syntax error",
		},
		{
			name: "Error with filename only",
			err: &ParseError{
				Filename: "test.apimock",
				Line:     0,
				Message:  "file error",
			},
			expected: "test.apimock: file error",
		},
		{
			name: "Error with line only",
			err: &ParseError{
				Filename: "",
				Line:     10,
				Message:  "parsing error",
			},
			expected: "line 10: parsing error",
		},
		{
			name: "Error with message only",
			err: &ParseError{
				Filename: "",
				Line:     0,
				Message:  "generic error",
			},
			expected: "generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "Error with field",
			err: &ValidationError{
				Field:   "StatusCode",
				Message: "must be between 100-599",
			},
			expected: "validation error in StatusCode: must be between 100-599",
		},
		{
			name: "Error without field",
			err: &ValidationError{
				Field:   "",
				Message: "general validation error",
			},
			expected: "validation error: general validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestAPIMockFile_Validate(t *testing.T) {
	t.Run("Valid file", func(t *testing.T) {
		file := &APIMockFile{
			Responses: []ResponseSection{
				{StatusCode: 200, Description: "OK"},
			},
		}
		err := file.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Empty responses", func(t *testing.T) {
		file := &APIMockFile{
			Responses: []ResponseSection{},
		}
		err := file.Validate()
		if err == nil {
			t.Error("expected error for empty responses")
		}
		if !strings.Contains(err.Error(), "at least one response") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Invalid request section", func(t *testing.T) {
		file := &APIMockFile{
			Request: &RequestSection{
				Method: "INVALID",
				Path:   "/test",
			},
			Responses: []ResponseSection{
				{StatusCode: 200},
			},
		}
		err := file.Validate()
		if err == nil {
			t.Error("expected error for invalid method")
		}
	})

	t.Run("Invalid response", func(t *testing.T) {
		file := &APIMockFile{
			Responses: []ResponseSection{
				{StatusCode: 999}, // Invalid status code
			},
		}
		err := file.Validate()
		if err == nil {
			t.Error("expected error for invalid status code")
		}
	})
}

func TestRequestSection_Validate(t *testing.T) {
	t.Run("Valid request with method", func(t *testing.T) {
		req := &RequestSection{
			Method: "POST",
			Path:   "/api/users",
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Valid request without method", func(t *testing.T) {
		req := &RequestSection{
			Method: "",
			Path:   "/api/users",
		}
		err := req.Validate()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("Empty path", func(t *testing.T) {
		req := &RequestSection{
			Method: "GET",
			Path:   "",
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for empty path")
		}
		if !strings.Contains(err.Error(), "path is required") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Invalid HTTP method", func(t *testing.T) {
		req := &RequestSection{
			Method: "INVALID",
			Path:   "/test",
		}
		err := req.Validate()
		if err == nil {
			t.Error("expected error for invalid method")
		}
		if !strings.Contains(err.Error(), "invalid HTTP method") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("Valid HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT"}
		for _, method := range methods {
			req := &RequestSection{
				Method: method,
				Path:   "/test",
			}
			err := req.Validate()
			if err != nil {
				t.Errorf("method %s should be valid, got error: %v", method, err)
			}
		}
	})
}

func TestResponseSection_Validate(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		shouldFail bool
	}{
		{"Valid 200", 200, false},
		{"Valid 404", 404, false},
		{"Valid 500", 500, false},
		{"Valid 100", 100, false},
		{"Valid 599", 599, false},
		{"Invalid 99", 99, true},
		{"Invalid 600", 600, true},
		{"Invalid 0", 0, true},
		{"Invalid -1", -1, true},
		{"Invalid 1000", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &ResponseSection{
				StatusCode: tt.statusCode,
			}
			err := resp.Validate()
			if tt.shouldFail && err == nil {
				t.Errorf("expected error for status code %d", tt.statusCode)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("unexpected error for status code %d: %v", tt.statusCode, err)
			}
			if tt.shouldFail && err != nil && !strings.Contains(err.Error(), "invalid HTTP status code") {
				t.Errorf("unexpected error message: %v", err)
			}
		})
	}
}
