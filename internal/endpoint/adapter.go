package endpoint

import (
	"fmt"
	"strings"

	"github.com/pretodev/anansi-proxy/pkg/apimock"
)

// FromAPIMockFile converts an APIMockFile to an EndpointSchema.
// This adapter allows the internal server and UI to work with the apimock package.
func FromAPIMockFile(ast *apimock.APIMockFile) (*EndpointSchema, error) {
	if len(ast.Responses) == 0 {
		return nil, fmt.Errorf("no responses found in APIMock file")
	}

	endpoint := &EndpointSchema{
		Route:     "/",
		Accept:    DefaultContentType,
		Responses: make([]Response, 0, len(ast.Responses)),
	}

	// Extract route from request section if available
	if ast.Request != nil {
		endpoint.Route = ast.Request.Path

		// Get Content-Type from request headers if available
		if contentType, ok := ast.Request.Headers["Content-Type"]; ok {
			endpoint.Accept = contentType
		}

		// Set body from request body schema if available
		if ast.Request.BodySchema != "" {
			endpoint.Body = ast.Request.BodySchema
		}
	}

	// Convert each response section
	for _, resp := range ast.Responses {
		response := Response{
			Title:       resp.Description,
			Body:        resp.Body,
			ContentType: DefaultContentType,
			StatusCode:  resp.StatusCode,
		}

		// If no description, create a default one
		if response.Title == "" {
			response.Title = fmt.Sprintf("Response %d", resp.StatusCode)
		}

		// Get Content-Type from response headers if available
		if contentType, ok := resp.Headers["Content-Type"]; ok {
			response.ContentType = contentType
		}

		endpoint.Responses = append(endpoint.Responses, response)
	}

	return endpoint, nil
}

// ParseAPIMock parses an .apimock file and converts it to an EndpointSchema.
// This is a convenience function that combines parsing and conversion.
func ParseAPIMock(filePath string) (*EndpointSchema, error) {
	// Create parser
	parser, err := apimock.NewParser(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser for '%s': %w", filePath, err)
	}

	// Parse the file
	ast, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse file '%s': %w", filePath, err)
	}

	// Validate the AST
	if err := ast.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed for '%s': %w", filePath, err)
	}

	// Convert to EndpointSchema
	endpoint, err := FromAPIMockFile(ast)
	if err != nil {
		return nil, fmt.Errorf("failed to convert APIMock file '%s': %w", filePath, err)
	}

	// Ensure route starts with /
	if !strings.HasPrefix(endpoint.Route, "/") {
		endpoint.Route = "/" + endpoint.Route
	}

	return endpoint, nil
}
