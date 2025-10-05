package endpoint

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	xsdvalidate "github.com/terminalstatic/go-xsd-validate"
)

type SchemaValidator interface {
	Validate(body string) error
}

func NewValidator(contentType string, schema string) (SchemaValidator, error) {
	if strings.TrimSpace(schema) == "" {
		return nil, nil
	}

	if contentType == "application/xml" {
		return NewXmlSchemaValidator(schema)
	}

	return NewJsonSchemaValidator(schema)
}

type JsonSchemaValidator struct {
	validator *jsonschema.Schema
}

func NewJsonSchemaValidator(schema string) (SchemaValidator, error) {
	var schemaDoc any
	if err := json.Unmarshal([]byte(schema), &schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	const schemaURL = "inmemory://schema.json"
	if err := compiler.AddResource(schemaURL, schemaDoc); err != nil {
		return nil, err
	}
	validator, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, err
	}
	return &JsonSchemaValidator{
		validator: validator,
	}, nil
}

// Validate implements SchemaValidator.
func (j *JsonSchemaValidator) Validate(body string) error {
	var data any
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if err := j.validator.Validate(data); err != nil {
		return fmt.Errorf("JSON validation failed: %w", err)
	}

	return nil
}

type XMLSchemaValidator struct {
	xsdHandler *xsdvalidate.XsdHandler
}

func NewXmlSchemaValidator(schema string) (SchemaValidator, error) {
	xsdvalidate.Init()
	xsdHandler, err := xsdvalidate.NewXsdHandlerMem([]byte(schema), xsdvalidate.ParsErrDefault)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XSD schema: %w", err)
	}

	return &XMLSchemaValidator{
		xsdHandler: xsdHandler,
	}, nil
}

func (x *XMLSchemaValidator) Validate(body string) error {
	if x.xsdHandler == nil {
		return fmt.Errorf("XSD handler not initialized")
	}

	if err := x.xsdHandler.ValidateMem([]byte(body), xsdvalidate.ValidErrDefault); err != nil {
		return fmt.Errorf("XML validation failed: %w", err)
	}

	return nil
}

// Free releases the resources held by the XMLSchemaValidator.
// This should be called when the validator is no longer needed.
func (x *XMLSchemaValidator) Free() {
	if x.xsdHandler != nil {
		x.xsdHandler.Free()
	}
}
