package apimock

import "fmt"

// ParseError represents an error that occurred during parsing.
// It includes the filename, line number, and error message for better debugging.
type ParseError struct {
	Filename string
	Line     int
	Message  string
}

// Error implements the error interface for ParseError.
func (e *ParseError) Error() string {
	if e.Filename != "" && e.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", e.Filename, e.Line, e.Message)
	}
	if e.Filename != "" {
		return fmt.Sprintf("%s: %s", e.Filename, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return e.Message
}

// NewParseError creates a new ParseError with the given details.
func NewParseError(filename string, line int, message string) *ParseError {
	return &ParseError{
		Filename: filename,
		Line:     line,
		Message:  message,
	}
}

// ValidationError represents an error found during AST validation.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface for ValidationError.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
