package apimock

import (
	"testing"
)

func TestEvaluator_BodyContains(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		search   string
		expected bool
	}{
		{
			name:     "Body contains email",
			body:     `{"email": "admin@example.com", "name": "Admin"}`,
			search:   "admin@example.com",
			expected: true,
		},
		{
			name:     "Body does not contain email",
			body:     `{"email": "user@example.com", "name": "User"}`,
			search:   "admin@example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqCtx := &RequestContext{
				Body: tt.body,
			}
			ctx := NewExecutionContext(1, reqCtx)
			ev := NewEvaluator(ctx)

			// Condition: body >> .contains "admin@example.com"
			condition := FunctionCall{
				Target: VariableReference{Name: "body"},
				Name:   "contains",
				Args: []Expression{
					StringValue{Value: tt.search},
				},
			}

			result, err := ev.Evaluate(condition)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			boolResult, ok := result.(bool)
			if !ok {
				t.Fatalf("Expected boolean result, got %T: %v", result, result)
			}

			if boolResult != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, boolResult)
			} else {
				t.Logf("✓ Correct: %v", boolResult)
			}
		})
	}
}
