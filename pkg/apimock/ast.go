package apimock

// APIMockFile represents the complete parsed .apimock file
type APIMockFile struct {
	Request   *RequestSection   // Optional request section
	Responses []ResponseSection // At least one response section
}

// RequestSection represents the HTTP request definition
type RequestSection struct {
	Method       string            // HTTP method (GET, POST, etc.) - optional
	Path         string            // Base path (e.g., "/api/users")
	PathSegments []PathSegment     // Parsed path segments with placeholders
	QueryParams  map[string]string // Query parameters
	Headers      map[string]string // HTTP headers
	BodySchema   string            // Request body schema (JSON, XML, etc.)
}

// PathSegment represents a segment in the URL path
type PathSegment struct {
	Value       string // The actual value
	IsParameter bool   // true if it's a placeholder like {id}
	Name        string // Parameter name (only if IsParameter is true)
}

// ResponseSection represents one HTTP response definition
type ResponseSection struct {
	StatusCode  int               // HTTP status code (200, 404, etc.)
	Description string            // Optional description
	Headers     map[string]string // Response headers
	Body        string            // Response body content
}

// NewAPIMockFile creates a new empty APIMock file structure
func NewAPIMockFile() *APIMockFile {
	return &APIMockFile{
		Responses: make([]ResponseSection, 0),
	}
}

// NewRequestSection creates a new empty request section
func NewRequestSection() *RequestSection {
	return &RequestSection{
		PathSegments: make([]PathSegment, 0),
		QueryParams:  make(map[string]string),
		Headers:      make(map[string]string),
	}
}

// NewResponseSection creates a new empty response section
func NewResponseSection() ResponseSection {
	return ResponseSection{
		Headers: make(map[string]string),
	}
}
