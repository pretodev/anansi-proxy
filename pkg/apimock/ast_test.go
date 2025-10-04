package apimock

import "testing"

func TestNewAPIMockFile(t *testing.T) {
	file := NewAPIMockFile()

	if file == nil {
		t.Fatal("NewAPIMockFile returned nil")
	}

	if file.Request != nil {
		t.Error("expected Request to be nil")
	}

	if file.Responses == nil {
		t.Error("expected Responses to be initialized")
	}

	if len(file.Responses) != 0 {
		t.Errorf("expected empty Responses slice, got length %d", len(file.Responses))
	}
}

func TestNewRequestSection(t *testing.T) {
	req := NewRequestSection()

	if req == nil {
		t.Fatal("NewRequestSection returned nil")
	}

	if req.PathSegments == nil {
		t.Error("expected PathSegments to be initialized")
	}

	if req.QueryParams == nil {
		t.Error("expected QueryParams to be initialized")
	}

	if req.Properties == nil {
		t.Error("expected Properties to be initialized")
	}

	if len(req.PathSegments) != 0 {
		t.Errorf("expected empty PathSegments, got length %d", len(req.PathSegments))
	}

	if len(req.QueryParams) != 0 {
		t.Errorf("expected empty QueryParams, got length %d", len(req.QueryParams))
	}

	if len(req.Properties) != 0 {
		t.Errorf("expected empty Properties, got length %d", len(req.Properties))
	}
}

func TestNewResponseSection(t *testing.T) {
	resp := NewResponseSection()

	if resp.Properties == nil {
		t.Error("expected Properties to be initialized")
	}

	if len(resp.Properties) != 0 {
		t.Errorf("expected empty Properties, got length %d", len(resp.Properties))
	}

	if resp.StatusCode != 0 {
		t.Errorf("expected StatusCode to be 0, got %d", resp.StatusCode)
	}

	if resp.Description != "" {
		t.Errorf("expected empty Description, got %s", resp.Description)
	}

	if resp.Body != "" {
		t.Errorf("expected empty Body, got %s", resp.Body)
	}
}

func TestPathSegment_Parameter(t *testing.T) {
	seg := PathSegment{
		Value:       "{id}",
		IsParameter: true,
		Name:        "id",
	}

	if !seg.IsParameter {
		t.Error("expected IsParameter to be true")
	}

	if seg.Name != "id" {
		t.Errorf("expected Name to be 'id', got '%s'", seg.Name)
	}

	if seg.Value != "{id}" {
		t.Errorf("expected Value to be '{id}', got '%s'", seg.Value)
	}
}

func TestPathSegment_Static(t *testing.T) {
	seg := PathSegment{
		Value:       "users",
		IsParameter: false,
		Name:        "",
	}

	if seg.IsParameter {
		t.Error("expected IsParameter to be false")
	}

	if seg.Name != "" {
		t.Errorf("expected empty Name, got '%s'", seg.Name)
	}

	if seg.Value != "users" {
		t.Errorf("expected Value to be 'users', got '%s'", seg.Value)
	}
}

func TestAPIMockFile_Structure(t *testing.T) {
	file := &APIMockFile{
		Request: &RequestSection{
			Method: "POST",
			Path:   "/api/users",
			PathSegments: []PathSegment{
				{Value: "api", IsParameter: false},
				{Value: "users", IsParameter: false},
			},
			QueryParams: map[string]string{
				"active": "true",
			},
			Properties: map[string]string{
				"Accept": "application/json",
			},
			BodySchema: `{"name": "test"}`,
		},
		Responses: []ResponseSection{
			{
				StatusCode:  201,
				Description: "Created",
				Properties: map[string]string{
					"Location": "/api/users/123",
				},
				Body: `{"id": 123}`,
			},
			{
				StatusCode:  400,
				Description: "Bad Request",
				Properties:  map[string]string{},
				Body:        `{"error": "Invalid input"}`,
			},
		},
	}

	// Verify request
	if file.Request.Method != "POST" {
		t.Errorf("expected method POST, got %s", file.Request.Method)
	}

	if len(file.Request.PathSegments) != 2 {
		t.Errorf("expected 2 path segments, got %d", len(file.Request.PathSegments))
	}

	if file.Request.QueryParams["active"] != "true" {
		t.Error("expected query param 'active=true'")
	}

	if file.Request.Properties["Accept"] != "application/json" {
		t.Error("expected Accept property")
	}

	// Verify responses
	if len(file.Responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(file.Responses))
	}

	if file.Responses[0].StatusCode != 201 {
		t.Errorf("expected first response status 201, got %d", file.Responses[0].StatusCode)
	}

	if file.Responses[1].StatusCode != 400 {
		t.Errorf("expected second response status 400, got %d", file.Responses[1].StatusCode)
	}
}

func TestRequestSection_GetPathParameters(t *testing.T) {
	req := &RequestSection{
		PathSegments: []PathSegment{
			{Value: "api", IsParameter: false},
			{Value: "{id}", IsParameter: true, Name: "id"},
			{Value: "posts", IsParameter: false},
			{Value: "{postId}", IsParameter: true, Name: "postId"},
		},
	}

	params := req.GetPathParameters()
	if len(params) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(params))
	}

	if params[0] != "id" {
		t.Errorf("expected first parameter to be 'id', got '%s'", params[0])
	}

	if params[1] != "postId" {
		t.Errorf("expected second parameter to be 'postId', got '%s'", params[1])
	}
}

func TestRequestSection_GetPathParameters_Empty(t *testing.T) {
	req := &RequestSection{
		PathSegments: []PathSegment{
			{Value: "api", IsParameter: false},
			{Value: "users", IsParameter: false},
		},
	}

	params := req.GetPathParameters()
	if len(params) != 0 {
		t.Errorf("expected no parameters, got %d", len(params))
	}
}

func TestRequestSection_HasPathParameters(t *testing.T) {
	tests := []struct {
		name     string
		segments []PathSegment
		expected bool
	}{
		{
			name: "Has parameters",
			segments: []PathSegment{
				{Value: "users", IsParameter: false},
				{Value: "{id}", IsParameter: true, Name: "id"},
			},
			expected: true,
		},
		{
			name: "No parameters",
			segments: []PathSegment{
				{Value: "api", IsParameter: false},
				{Value: "users", IsParameter: false},
			},
			expected: false,
		},
		{
			name:     "Empty segments",
			segments: []PathSegment{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RequestSection{PathSegments: tt.segments}
			got := req.HasPathParameters()
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestPathSegment_String(t *testing.T) {
	tests := []struct {
		segment  PathSegment
		expected string
	}{
		{
			segment:  PathSegment{Value: "users", IsParameter: false},
			expected: "users",
		},
		{
			segment:  PathSegment{Value: "{id}", IsParameter: true, Name: "id"},
			expected: "{id}",
		},
	}

	for _, tt := range tests {
		got := tt.segment.String()
		if got != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, got)
		}
	}
}

func TestIsValidHTTPMethod(t *testing.T) {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "TRACE", "CONNECT"}
	for _, method := range validMethods {
		if !IsValidHTTPMethod(method) {
			t.Errorf("expected %s to be valid", method)
		}
	}

	invalidMethods := []string{"INVALID", "get", "post", "UNKNOWN", ""}
	for _, method := range invalidMethods {
		if IsValidHTTPMethod(method) {
			t.Errorf("expected %s to be invalid", method)
		}
	}
}

func TestIsValidHTTPStatusCode(t *testing.T) {
	validCodes := []int{100, 200, 201, 404, 500, 599}
	for _, code := range validCodes {
		if !IsValidHTTPStatusCode(code) {
			t.Errorf("expected %d to be valid", code)
		}
	}

	invalidCodes := []int{0, 99, 600, 1000, -1}
	for _, code := range invalidCodes {
		if IsValidHTTPStatusCode(code) {
			t.Errorf("expected %d to be invalid", code)
		}
	}
}
