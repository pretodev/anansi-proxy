// Package apimock provides parsing and lexical analysis for .apimock files.
// APIMock files define HTTP API mock specifications including requests and responses.
package apimock

import (
	"fmt"
	"strings"
)

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
	Properties   map[string]string // Request Properties
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

// ============================================
// Conditions AST - Represents conditional logic
// ============================================

// ConditionLine represents a single condition line in a response.
// Multiple condition lines are combined with AND logic by default,
// or OR logic if IsOrCondition is true.
type ConditionLine struct {
	Expression    Expression // The condition expression to evaluate
	IsOrCondition bool       // true if this line starts with "or" keyword
	Line          int        // Source line number for error reporting
}

// Expression represents a complete conditional expression.
// It can be a simple value or a complex expression with operators.
type Expression interface {
	// String returns a string representation of the expression
	String() string
}

// BinaryExpression represents an expression with a binary operator (e.g., a + b, x == y).
type BinaryExpression struct {
	Left     Expression
	Operator string // Operators: +, -, *, /, %, //, .., ==, !=, >, <, >=, <=, and, or
	Right    Expression
}

func (e BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", e.Left.String(), e.Operator, e.Right.String())
}

// UnaryExpression represents an expression with a unary operator (e.g., not x).
type UnaryExpression struct {
	Operator string // Operator: not
	Operand  Expression
}

func (e UnaryExpression) String() string {
	return fmt.Sprintf("(%s %s)", e.Operator, e.Operand.String())
}

// Attribution represents variable assignment with the >> operator.
// Supports destructuring: value >> var1, var2, var3
type Attribution struct {
	Value     Expression
	Variables []string // Variable names to assign to (supports destructuring)
}

func (a Attribution) String() string {
	if len(a.Variables) == 1 {
		return fmt.Sprintf("(%s >> %s)", a.Value.String(), a.Variables[0])
	}
	return fmt.Sprintf("(%s >> %s)", a.Value.String(), strings.Join(a.Variables, ", "))
}

// Value represents primitive values and complex data structures.
type Value interface {
	Expression
	// GetValue returns the underlying value
	GetValue() interface{}
}

// NumberValue represents a numeric literal (integer or float).
type NumberValue struct {
	Value float64
}

func (n NumberValue) String() string {
	// Format number without decimal if it's an integer
	if n.Value == float64(int(n.Value)) {
		return fmt.Sprintf("%d", int(n.Value))
	}
	return fmt.Sprintf("%g", n.Value)
}

func (n NumberValue) GetValue() interface{} {
	return n.Value
}

// BooleanValue represents a boolean literal (True or False).
type BooleanValue struct {
	Value bool
}

func (b BooleanValue) String() string {
	if b.Value {
		return "True"
	}
	return "False"
}

func (b BooleanValue) GetValue() interface{} {
	return b.Value
}

// StringValue represents a string literal.
type StringValue struct {
	Value string
}

func (s StringValue) String() string {
	return fmt.Sprintf(`"%s"`, s.Value)
}

func (s StringValue) GetValue() interface{} {
	return s.Value
}

// TableValue represents a table (array or dictionary).
type TableValue struct {
	IsArray bool             // true for array-style, false for dict-style
	Array   []Value          // Array elements
	Dict    map[string]Value // Dictionary key-value pairs
}

func (t TableValue) String() string {
	if t.IsArray {
		if len(t.Array) == 0 {
			return "[]"
		}
		parts := make([]string, len(t.Array))
		for i, v := range t.Array {
			parts[i] = v.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	// Dictionary
	if len(t.Dict) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(t.Dict))
	for k, v := range t.Dict {
		parts = append(parts, fmt.Sprintf("%s = %s", k, v.String()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func (t TableValue) GetValue() interface{} {
	if t.IsArray {
		return t.Array
	}
	return t.Dict
}

// RangeValue represents a numeric range (e.g., 1..10).
type RangeValue struct {
	Start float64
	End   float64
}

func (r RangeValue) String() string {
	// Format numbers without decimal if they're integers
	if r.Start == float64(int(r.Start)) && r.End == float64(int(r.End)) {
		return fmt.Sprintf("%d..%d", int(r.Start), int(r.End))
	}
	return fmt.Sprintf("%g..%g", r.Start, r.End)
}

func (r RangeValue) GetValue() interface{} {
	return r
}

// VariableReference represents a reference to a variable or context value.
// Examples: call_count, body.email, headers["Authorization"]
type VariableReference struct {
	Name       string   // Base variable name (e.g., "body", "headers")
	AccessPath []Access // Chain of property/index accesses
}

// Access represents a single property or index access.
type Access struct {
	Type AccessType // Property (dot notation) or Index (bracket notation)
	Key  string     // Property name or index key
}

// AccessType defines how a value is accessed.
type AccessType int

const (
	PropertyAccess AccessType = iota // Dot notation: obj.property
	IndexAccess                      // Bracket notation: obj["key"]
)

func (v VariableReference) String() string {
	s := v.Name
	for _, access := range v.AccessPath {
		if access.Type == PropertyAccess {
			s += "." + access.Key
		} else {
			s += "[\"" + access.Key + "\"]"
		}
	}
	return s
}

func (v VariableReference) GetValue() interface{} {
	return v
}

// FunctionCall represents a built-in function call.
// Functions start with dot (e.g., .split, .contains, .random_int).
type FunctionCall struct {
	Target Expression   // The value to operate on (can be nil for global functions)
	Name   string       // Function name without the dot
	Args   []Expression // Function arguments
}

func (f FunctionCall) String() string {
	// Format arguments
	argStrs := make([]string, len(f.Args))
	for i, arg := range f.Args {
		argStrs[i] = arg.String()
	}
	argList := strings.Join(argStrs, ", ")

	if f.Target != nil {
		return fmt.Sprintf("%s.%s(%s)", f.Target.String(), f.Name, argList)
	}
	return fmt.Sprintf(".%s(%s)", f.Name, argList)
}

func (f FunctionCall) GetValue() interface{} {
	return f
}

// ResponseSection represents one HTTP response definition.
// Each response includes a status code, optional description,
// properties, optional conditions, and response body content.
type ResponseSection struct {
	StatusCode  int               // HTTP status code (200, 404, etc.)
	Description string            // Optional description
	Properties  map[string]string // Response Properties
	Conditions  []ConditionLine   // Optional condition lines
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
		Properties:   make(map[string]string),
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
// The Properties map and Conditions slice are initialized to empty.
func NewResponseSection() ResponseSection {
	return ResponseSection{
		Properties: make(map[string]string),
		Conditions: make([]ConditionLine, 0),
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
	
	// Validate conditions
	if len(r.Conditions) > 0 {
		validator := NewConditionValidator()
		if err := validator.ValidateConditions(r.Conditions); err != nil {
			return NewValidationError("Conditions", err.Error())
		}
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
