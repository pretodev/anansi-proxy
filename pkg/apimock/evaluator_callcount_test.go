package apimock

import (
	"fmt"
	"testing"
)

func TestEvaluator_CallCountCondition(t *testing.T) {
	tests := []struct {
		callCount int
		expected  bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, true},
		{5, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("call_count=%d", tt.callCount), func(t *testing.T) {
			ctx := NewExecutionContext(tt.callCount, &RequestContext{})
			ev := NewEvaluator(ctx)

			// Condition: call_count > 3
			condition := BinaryExpression{
				Left:     VariableReference{Name: "call_count"},
				Operator: ">",
				Right:    NumberValue{Value: 3},
			}

			result, err := ev.Evaluate(condition)
			if err != nil {
				t.Fatalf("Evaluation error: %v", err)
			}

			boolResult, ok := result.(bool)
			if !ok {
				t.Fatalf("Expected boolean result, got %T", result)
			}

			if boolResult != tt.expected {
				t.Errorf("call_count=%d: expected %v, got %v", tt.callCount, tt.expected, boolResult)
			}
		})
	}
}
