package endpoint

import (
	"fmt"
	"strings"

	"github.com/pretodev/anansi-proxy/pkg/apimock"
)

// EndpointWithFile represents an endpoint schema along with its source file
type EndpointWithFile struct {
	Schema   *EndpointSchema
	FilePath string
}

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

	if ast.Request != nil {
		method := ""
		if ast.Request.Method != "" {
			method = strings.ToUpper(ast.Request.Method) + " "
		}

		endpoint.Route = method + ast.Request.Path

		if contentType, ok := ast.Request.Properties[RequestAcceptPropertyName]; ok {
			endpoint.Accept = contentType
		}

		if ast.Request.BodySchema != "" {
			endpoint.Body = ast.Request.BodySchema
			validator, err := NewValidator(endpoint.Accept, endpoint.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to create schema validator: %w", err)
			}
			endpoint.Validator = validator
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

		if contentType, ok := resp.Properties[ResponseContentTypePropertyName]; ok {
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

	return endpoint, nil
}

// ParseAPIMockFiles parses multiple .apimock files and returns a slice of EndpointWithFile.
// This function processes each file and collects all successfully parsed endpoints.
func ParseAPIMockFiles(filePaths ...string) ([]*EndpointWithFile, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no file paths provided")
	}

	endpoints := make([]*EndpointWithFile, 0, len(filePaths))
	var errors []string

	for _, filePath := range filePaths {
		endpoint, err := ParseAPIMock(filePath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("- %s: %v", filePath, err))
			continue
		}

		endpoints = append(endpoints, &EndpointWithFile{
			Schema:   endpoint,
			FilePath: filePath,
		})
	}

	if len(errors) > 0 {
		if len(endpoints) == 0 {
			return nil, fmt.Errorf("failed to parse all files:\n%s", strings.Join(errors, "\n"))
		}
		// Log warnings but continue if we have at least some valid endpoints
		fmt.Printf("Warning: some files failed to parse:\n%s\n", strings.Join(errors, "\n"))
	}

	return endpoints, nil
}
