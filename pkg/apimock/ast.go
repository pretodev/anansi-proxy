// Package apimock provides parsing and lexical analysis for .apimock files.
// APIMock files define HTTP API mock specifications including requests and responses.
package apimock

import "fmt"

// HTTP method constants
const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
	MethodConnect = "CONNECT"
)

// HTTP status code ranges
const (
	MinHTTPStatusCode = 100
	MaxHTTPStatusCode = 599
)

// validHTTPMethods contains all valid HTTP methods
var validHTTPMethods = map[string]bool{
	MethodGet:     true,
	MethodPost:    true,
	MethodPut:     true,
	MethodDelete:  true,
	MethodPatch:   true,
	MethodHead:    true,
	MethodOptions: true,
	MethodTrace:   true,
	MethodConnect: true,
}

// APIMockFile represents the complete parsed .apimock file.
// It contains an optional request section and one or more response sections.
type APIMockFile struct {
	Request   *RequestSection   // Optional request section
	Responses []ResponseSection // At least one response section
}

// RequestSection represents the HTTP request definition.
// It includes the HTTP method, path with parameters, query parameters,
// headers, and an optional request body schema.
type RequestSection struct {
	Method       string            // HTTP method (GET, POST, etc.) - optional
	Path         string            // Base path (e.g., "/api/users")
	PathSegments []PathSegment     // Parsed path segments with placeholders
	QueryParams  map[string]string // Query parameters
	Headers      map[string]string // HTTP headers
	BodySchema   string            // Request body schema (JSON, XML, etc.)
}

// PathSegment represents a segment in the URL path.
// A segment can be either a static value (e.g., "users") or a parameter
// placeholder (e.g., "{id}").
type PathSegment struct {
	Value       string // The actual value
	IsParameter bool   // true if it's a placeholder like {id}
	Name        string // Parameter name (only if IsParameter is true)
}

// String returns the string representation of the path segment.
func (ps PathSegment) String() string {
	return ps.Value
}

// ResponseSection represents one HTTP response definition.
// Each response includes a status code, optional description,
// headers, and response body content.
type ResponseSection struct {
	StatusCode  int               // HTTP status code (200, 404, etc.)
	Description string            // Optional description
	Headers     map[string]string // Response headers
	Body        string            // Response body content
}

// NewAPIMockFile creates a new empty APIMock file structure.
// The returned file has no request section and an empty responses slice.
func NewAPIMockFile() *APIMockFile {
	return &APIMockFile{
		Responses: make([]ResponseSection, 0),
	}
}

// NewRequestSection creates a new empty request section.
// All maps (QueryParams, Headers) are initialized to empty maps.
func NewRequestSection() *RequestSection {
	return &RequestSection{
		PathSegments: make([]PathSegment, 0),
		QueryParams:  make(map[string]string),
		Headers:      make(map[string]string),
	}
}

// GetPathParameters returns a slice of all path parameter names.
// For example, from "/users/{id}/posts/{postId}" it returns ["id", "postId"].
func (r *RequestSection) GetPathParameters() []string {
	params := make([]string, 0)
	for _, seg := range r.PathSegments {
		if seg.IsParameter {
			params = append(params, seg.Name)
		}
	}
	return params
}

// HasPathParameters returns true if the request path contains any parameters.
func (r *RequestSection) HasPathParameters() bool {
	for _, seg := range r.PathSegments {
		if seg.IsParameter {
			return true
		}
	}
	return false
}

// NewResponseSection creates a new empty response section.
// The Headers map is initialized to an empty map.
func NewResponseSection() ResponseSection {
	return ResponseSection{
		Headers: make(map[string]string),
	}
}

// Validate checks if the APIMockFile is semantically valid.
// It verifies that all required fields are properly set and that
// values are within acceptable ranges.
func (f *APIMockFile) Validate() error {
	if len(f.Responses) == 0 {
		return NewValidationError("Responses", "at least one response section is required")
	}

	// Validate request section if present
	if f.Request != nil {
		if err := f.Request.Validate(); err != nil {
			return err
		}
	}

	// Validate all responses
	for i, resp := range f.Responses {
		if err := resp.Validate(); err != nil {
			return fmt.Errorf("response %d: %w", i, err)
		}
	}

	return nil
}

// Validate checks if the RequestSection is semantically valid.
func (r *RequestSection) Validate() error {
	if r.Path == "" {
		return NewValidationError("Path", "path is required")
	}

	// Validate HTTP method if present
	if r.Method != "" && !IsValidHTTPMethod(r.Method) {
		return NewValidationError("Method", fmt.Sprintf("invalid HTTP method: %s", r.Method))
	}

	return nil
}

// Validate checks if the ResponseSection is semantically valid.
func (r *ResponseSection) Validate() error {
	if !IsValidHTTPStatusCode(r.StatusCode) {
		return NewValidationError("StatusCode", fmt.Sprintf("invalid HTTP status code: %d (must be between %d-%d)", r.StatusCode, MinHTTPStatusCode, MaxHTTPStatusCode))
	}
	return nil
}

// IsValidHTTPMethod checks if the given method is a valid HTTP method.
func IsValidHTTPMethod(method string) bool {
	return validHTTPMethods[method]
}

// IsValidHTTPStatusCode checks if the given status code is within the valid range.
func IsValidHTTPStatusCode(code int) bool {
	return code >= MinHTTPStatusCode && code <= MaxHTTPStatusCode
}
