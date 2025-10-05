package endpoint

import (
	"fmt"
	"strings"
	"testing"
)

func TestJsonSchemaValidator_Validate(t *testing.T) {
	schema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "number"}
		},
		"required": ["name"]
	}`

	validator, err := NewJsonSchemaValidator(schema)
	if err != nil {
		t.Fatalf("Failed to create JSON schema validator: %v", err)
	}

	tests := []struct {
		name      string
		body      string
		shouldErr bool
	}{
		{
			name:      "Valid JSON",
			body:      `{"name": "John", "age": 30}`,
			shouldErr: false,
		},
		{
			name:      "Valid JSON without optional field",
			body:      `{"name": "Jane"}`,
			shouldErr: false,
		},
		{
			name:      "Missing required field",
			body:      `{"age": 25}`,
			shouldErr: true,
		},
		{
			name:      "Wrong type",
			body:      `{"name": 123}`,
			shouldErr: true,
		},
		{
			name:      "Invalid JSON",
			body:      `{invalid}`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.body)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Validate() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestXMLSchemaValidator_Validate(t *testing.T) {
	schema := `<?xml version="1.0" encoding="UTF-8"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="person">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="name" type="xs:string"/>
        <xs:element name="age" type="xs:integer"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

	validator, err := NewXmlSchemaValidator(schema)
	if err != nil {
		t.Fatalf("Failed to create XML schema validator: %v", err)
	}
	defer func() {
		if v, ok := validator.(*XMLSchemaValidator); ok {
			v.Free()
		}
	}()

	tests := []struct {
		name      string
		body      string
		shouldErr bool
	}{
		{
			name: "Valid XML",
			body: `<?xml version="1.0" encoding="UTF-8"?>
<person>
  <name>John Doe</name>
  <age>30</age>
</person>`,
			shouldErr: false,
		},
		{
			name: "Missing required element",
			body: `<?xml version="1.0" encoding="UTF-8"?>
<person>
  <name>Jane Doe</name>
</person>`,
			shouldErr: true,
		},
		{
			name: "Invalid element type",
			body: `<?xml version="1.0" encoding="UTF-8"?>
<person>
  <name>Bob</name>
  <age>not a number</age>
</person>`,
			shouldErr: true,
		},
		{
			name:      "Invalid XML syntax",
			body:      `<person><name>Invalid</person>`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.body)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Validate() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		schema      string
		wantType    string
		wantNil     bool
	}{
		{
			name:        "JSON validator",
			contentType: "application/json",
			schema:      `{"type": "object"}`,
			wantType:    "*endpoint.JsonSchemaValidator",
			wantNil:     false,
		},
		{
			name:        "XML validator",
			contentType: "application/xml",
			schema: `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root" type="xs:string"/>
</xs:schema>`,
			wantType: "*endpoint.XMLSchemaValidator",
			wantNil:  false,
		},
		{
			name:        "Empty schema returns nil",
			contentType: "application/json",
			schema:      "",
			wantNil:     true,
		},
		{
			name:        "Whitespace schema returns nil",
			contentType: "application/json",
			schema:      "   \n\t  ",
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := NewValidator(tt.contentType, tt.schema)

			if tt.wantNil {
				if validator != nil {
					t.Errorf("Expected nil validator, got %T", validator)
				}
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewValidator() error = %v", err)
			}

			if validator == nil {
				t.Fatal("Expected validator, got nil")
			}

			gotType := strings.TrimPrefix(strings.TrimPrefix(fmt.Sprintf("%T", validator), "*"), "endpoint.")
			wantTypeShort := strings.TrimPrefix(tt.wantType, "*endpoint.")

			if gotType != wantTypeShort {
				t.Errorf("Expected validator type %s, got %s", wantTypeShort, gotType)
			}

			// Cleanup XML validator
			if v, ok := validator.(*XMLSchemaValidator); ok {
				v.Free()
			}
		})
	}
}
