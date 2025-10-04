package server

import (
	"encoding/json"
	"strings"
)

// TODO: Implement full JSON schema validation
func validateRequestBody(body []byte, schema string) bool {
	if strings.Contains(schema, `"$schema"`) {
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err != nil {
			return false
		}
		// For now, just validate that it's valid JSON
		// TODO: Implement full JSON schema validation
		return true
	}

	// For XML schemas, just check if body is not empty
	if strings.Contains(schema, "xs:schema") || strings.Contains(schema, "xsd:schema") {
		return len(strings.TrimSpace(string(body))) > 0
	}

	// If no specific schema format detected, assume valid
	return true
}
