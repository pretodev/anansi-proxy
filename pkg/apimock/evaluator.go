package apimock

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ExecutionContext holds runtime state for condition evaluation.
type ExecutionContext struct {
	CallCount int                    // Number of times this endpoint has been called
	Request   *RequestContext        // Current request information
	Variables map[string]interface{} // Variables defined via >> operator
	rand      *rand.Rand             // Random number generator
}

// RequestContext contains information about the current HTTP request.
type RequestContext struct {
	Method  string            // HTTP method (GET, POST, etc.)
	Path    string            // Request path
	Query   map[string]string // Query parameters
	Headers map[string]string // Request headers
	Body    interface{}       // Parsed request body (can be string, map, array, etc.)
}

// NewExecutionContext creates a new execution context.
func NewExecutionContext(callCount int, request *RequestContext) *ExecutionContext {
	return &ExecutionContext{
		CallCount: callCount,
		Request:   request,
		Variables: make(map[string]interface{}),
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Evaluator evaluates condition expressions in an execution context.
type Evaluator struct {
	context *ExecutionContext
}

// NewEvaluator creates a new evaluator with the given context.
func NewEvaluator(context *ExecutionContext) *Evaluator {
	return &Evaluator{
		context: context,
	}
}

// EvaluateConditions evaluates a list of condition lines and returns true if all conditions pass.
// Conditions are implicitly AND'd together unless marked as OR conditions.
func (e *Evaluator) EvaluateConditions(conditions []ConditionLine) (bool, error) {
	if len(conditions) == 0 {
		return true, nil
	}

	result := true
	for _, cond := range conditions {
		value, err := e.Evaluate(cond.Expression)
		if err != nil {
			return false, err
		}

		boolValue, ok := value.(bool)
		if !ok {
			return false, fmt.Errorf("condition must evaluate to boolean, got %T", value)
		}

		if cond.IsOrCondition {
			// OR condition: if any OR condition is true, the whole group is true
			if boolValue {
				result = true
			}
		} else {
			// AND condition (default): all must be true
			result = result && boolValue
			if !result {
				// Short-circuit: if any AND condition fails, we're done
				return false, nil
			}
		}
	}

	return result, nil
}

// Evaluate evaluates an expression and returns its value.
func (e *Evaluator) Evaluate(expr Expression) (interface{}, error) {
	switch v := expr.(type) {
	case BinaryExpression:
		return e.evaluateBinaryExpression(v)
	case UnaryExpression:
		return e.evaluateUnaryExpression(v)
	case Attribution:
		return e.evaluateAttribution(v)
	case NumberValue:
		return v.Value, nil
	case BooleanValue:
		return v.Value, nil
	case StringValue:
		return v.Value, nil
	case TableValue:
		return e.evaluateTableValue(v)
	case RangeValue:
		return v, nil
	case VariableReference:
		return e.evaluateVariableReference(v)
	case FunctionCall:
		return e.evaluateFunctionCall(v)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evaluateBinaryExpression evaluates binary operators.
func (e *Evaluator) evaluateBinaryExpression(expr BinaryExpression) (interface{}, error) {
	left, err := e.Evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := e.Evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	// Arithmetic operators
	case "+":
		return e.add(left, right)
	case "-":
		return e.subtract(left, right)
	case "*":
		return e.multiply(left, right)
	case "/":
		return e.divide(left, right)
	case "%":
		return e.modulo(left, right)
	case "//":
		return e.integerDivide(left, right)

	// Comparison operators
	case "==":
		return e.equals(left, right), nil
	case "!=":
		return !e.equals(left, right), nil
	case ">":
		return e.greaterThan(left, right)
	case "<":
		return e.lessThan(left, right)
	case ">=":
		return e.greaterThanOrEqual(left, right)
	case "<=":
		return e.lessThanOrEqual(left, right)

	// String concatenation
	case "..":
		return fmt.Sprintf("%v%v", left, right), nil

	// Logical operators
	case "and":
		leftBool, lok := left.(bool)
		rightBool, rok := right.(bool)
		if !lok || !rok {
			return nil, fmt.Errorf("'and' operator requires boolean operands")
		}
		return leftBool && rightBool, nil

	case "or":
		leftBool, lok := left.(bool)
		rightBool, rok := right.(bool)
		if !lok || !rok {
			return nil, fmt.Errorf("'or' operator requires boolean operands")
		}
		return leftBool || rightBool, nil

	default:
		return nil, fmt.Errorf("unknown binary operator: %s", expr.Operator)
	}
}

// evaluateUnaryExpression evaluates unary operators.
func (e *Evaluator) evaluateUnaryExpression(expr UnaryExpression) (interface{}, error) {
	operand, err := e.Evaluate(expr.Operand)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "not":
		boolValue, ok := operand.(bool)
		if !ok {
			return nil, fmt.Errorf("'not' operator requires boolean operand, got %T", operand)
		}
		return !boolValue, nil

	default:
		return nil, fmt.Errorf("unknown unary operator: %s", expr.Operator)
	}
}

// evaluateAttribution evaluates variable attribution (>>).
func (e *Evaluator) evaluateAttribution(expr Attribution) (interface{}, error) {
	value, err := e.Evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	// Handle destructuring
	if len(expr.Variables) == 1 {
		// Single variable assignment
		e.context.Variables[expr.Variables[0]] = value
	} else {
		// Destructuring: try to extract values from table/map
		switch v := value.(type) {
		case map[string]interface{}:
			// Dictionary destructuring
			for _, varName := range expr.Variables {
				if val, ok := v[varName]; ok {
					e.context.Variables[varName] = val
				}
			}
		case []interface{}:
			// Array destructuring
			for i, varName := range expr.Variables {
				if i < len(v) {
					e.context.Variables[varName] = v[i]
				}
			}
		default:
			return nil, fmt.Errorf("cannot destructure value of type %T", value)
		}
	}

	return value, nil
}

// evaluateTableValue evaluates table (array or dictionary) values.
func (e *Evaluator) evaluateTableValue(table TableValue) (interface{}, error) {
	if table.IsArray {
		// Evaluate array elements
		result := make([]interface{}, len(table.Array))
		for i, elem := range table.Array {
			val, err := e.Evaluate(elem)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	}

	// Evaluate dictionary values
	result := make(map[string]interface{})
	for key, val := range table.Dict {
		evalVal, err := e.Evaluate(val)
		if err != nil {
			return nil, err
		}
		result[key] = evalVal
	}
	return result, nil
}

// evaluateVariableReference evaluates variable references with property/index access.
func (e *Evaluator) evaluateVariableReference(ref VariableReference) (interface{}, error) {
	// Get base variable
	var value interface{}

	switch ref.Name {
	case "call_count":
		value = e.context.CallCount
	case "request":
		value = e.context.Request
	default:
		// Check user-defined variables
		var ok bool
		value, ok = e.context.Variables[ref.Name]
		if !ok {
			return nil, fmt.Errorf("undefined variable: %s", ref.Name)
		}
	}

	// Apply access path
	for _, access := range ref.AccessPath {
		switch access.Type {
		case PropertyAccess:
			value = e.getProperty(value, access.Key)
		case IndexAccess:
			value = e.getIndex(value, access.Key)
		}

		if value == nil {
			return nil, fmt.Errorf("cannot access %s on nil value", access.Key)
		}
	}

	return value, nil
}

// getProperty accesses a property by name.
func (e *Evaluator) getProperty(obj interface{}, key string) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		return v[key]
	case map[string]string:
		return v[key]
	case *RequestContext:
		// Special handling for RequestContext
		switch key {
		case "method":
			return v.Method
		case "path":
			return v.Path
		case "query":
			return v.Query
		case "headers":
			return v.Headers
		case "body":
			return v.Body
		default:
			return nil
		}
	default:
		return nil
	}
}

// getIndex accesses an element by index (string or number).
func (e *Evaluator) getIndex(obj interface{}, key string) interface{} {
	switch v := obj.(type) {
	case map[string]interface{}:
		return v[key]
	case map[string]string:
		return v[key]
	case []interface{}:
		// Try to parse key as integer
		if idx, err := strconv.Atoi(key); err == nil && idx >= 0 && idx < len(v) {
			return v[idx]
		}
		return nil
	default:
		return nil
	}
}

// evaluateFunctionCall evaluates built-in function calls.
func (e *Evaluator) evaluateFunctionCall(fc FunctionCall) (interface{}, error) {
	// Evaluate target if present
	var target interface{}
	var err error
	if fc.Target != nil {
		target, err = e.Evaluate(fc.Target)
		if err != nil {
			return nil, err
		}
	}

	// Evaluate arguments
	args := make([]interface{}, len(fc.Args))
	for i, arg := range fc.Args {
		args[i], err = e.Evaluate(arg)
		if err != nil {
			return nil, err
		}
	}

	// Call built-in function
	return e.callBuiltinFunction(fc.Name, target, args)
}

// callBuiltinFunction calls a built-in function by name.
func (e *Evaluator) callBuiltinFunction(name string, target interface{}, args []interface{}) (interface{}, error) {
	switch name {
	// String functions
	case "split":
		return e.fnSplit(target, args)
	case "contains":
		return e.fnContains(target, args)
	case "matches":
		return e.fnMatches(target, args)
	case "upper":
		return e.fnUpper(target, args)
	case "lower":
		return e.fnLower(target, args)
	case "trim":
		return e.fnTrim(target, args)
	case "substring":
		return e.fnSubstring(target, args)

	// Collection functions
	case "len":
		return e.fnLen(target, args)

	// Random functions
	case "random":
		return e.fnRandom(target, args)
	case "random_int":
		return e.fnRandomInt(target, args)
	case "random_float":
		return e.fnRandomFloat(target, args)
	case "random_choice":
		return e.fnRandomChoice(target, args)

	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

// Arithmetic helper functions

func (e *Evaluator) add(a, b interface{}) (interface{}, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return nil, fmt.Errorf("'+' operator requires numeric operands")
	}
	return aNum + bNum, nil
}

func (e *Evaluator) subtract(a, b interface{}) (interface{}, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return nil, fmt.Errorf("'-' operator requires numeric operands")
	}
	return aNum - bNum, nil
}

func (e *Evaluator) multiply(a, b interface{}) (interface{}, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return nil, fmt.Errorf("'*' operator requires numeric operands")
	}
	return aNum * bNum, nil
}

func (e *Evaluator) divide(a, b interface{}) (interface{}, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return nil, fmt.Errorf("'/' operator requires numeric operands")
	}
	if bNum == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return aNum / bNum, nil
}

func (e *Evaluator) modulo(a, b interface{}) (interface{}, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return nil, fmt.Errorf("'%%' operator requires numeric operands")
	}
	if bNum == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return math.Mod(aNum, bNum), nil
}

func (e *Evaluator) integerDivide(a, b interface{}) (interface{}, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return nil, fmt.Errorf("'//' operator requires numeric operands")
	}
	if bNum == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return math.Floor(aNum / bNum), nil
}

// Comparison helper functions

func (e *Evaluator) equals(a, b interface{}) bool {
	// Handle different types
	switch aVal := a.(type) {
	case float64:
		if bVal, ok := toFloat64(b); ok {
			return aVal == bVal
		}
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	}
	return false
}

func (e *Evaluator) greaterThan(a, b interface{}) (bool, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("'>' operator requires numeric operands")
	}
	return aNum > bNum, nil
}

func (e *Evaluator) lessThan(a, b interface{}) (bool, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("'<' operator requires numeric operands")
	}
	return aNum < bNum, nil
}

func (e *Evaluator) greaterThanOrEqual(a, b interface{}) (bool, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("'>=' operator requires numeric operands")
	}
	return aNum >= bNum, nil
}

func (e *Evaluator) lessThanOrEqual(a, b interface{}) (bool, error) {
	aNum, aOk := toFloat64(a)
	bNum, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("'<=' operator requires numeric operands")
	}
	return aNum <= bNum, nil
}

// Built-in function implementations

// fnSplit splits a string by delimiter.
func (e *Evaluator) fnSplit(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("split() requires string target, got %T", target)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("split() requires 1 argument, got %d", len(args))
	}
	delimiter, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("split() delimiter must be string, got %T", args[0])
	}

	parts := strings.Split(str, delimiter)
	result := make([]interface{}, len(parts))
	for i, p := range parts {
		result[i] = p
	}
	return result, nil
}

// fnContains checks if a string contains a substring.
func (e *Evaluator) fnContains(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("contains() requires string target, got %T", target)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("contains() requires 1 argument, got %d", len(args))
	}
	substr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("contains() argument must be string, got %T", args[0])
	}

	return strings.Contains(str, substr), nil
}

// fnMatches checks if a string matches a regex pattern.
func (e *Evaluator) fnMatches(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("matches() requires string target, got %T", target)
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("matches() requires 1 argument, got %d", len(args))
	}
	pattern, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("matches() pattern must be string, got %T", args[0])
	}

	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %v", err)
	}
	return matched, nil
}

// fnUpper converts string to uppercase.
func (e *Evaluator) fnUpper(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("upper() requires string target, got %T", target)
	}
	if len(args) != 0 {
		return nil, fmt.Errorf("upper() takes no arguments, got %d", len(args))
	}
	return strings.ToUpper(str), nil
}

// fnLower converts string to lowercase.
func (e *Evaluator) fnLower(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("lower() requires string target, got %T", target)
	}
	if len(args) != 0 {
		return nil, fmt.Errorf("lower() takes no arguments, got %d", len(args))
	}
	return strings.ToLower(str), nil
}

// fnTrim trims whitespace from both ends of a string.
func (e *Evaluator) fnTrim(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("trim() requires string target, got %T", target)
	}
	if len(args) != 0 {
		return nil, fmt.Errorf("trim() takes no arguments, got %d", len(args))
	}
	return strings.TrimSpace(str), nil
}

// fnSubstring extracts a substring.
func (e *Evaluator) fnSubstring(target interface{}, args []interface{}) (interface{}, error) {
	str, ok := target.(string)
	if !ok {
		return nil, fmt.Errorf("substring() requires string target, got %T", target)
	}
	if len(args) != 2 {
		return nil, fmt.Errorf("substring() requires 2 arguments, got %d", len(args))
	}

	start, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("substring() start must be number, got %T", args[0])
	}
	end, ok := toFloat64(args[1])
	if !ok {
		return nil, fmt.Errorf("substring() end must be number, got %T", args[1])
	}

	iStart := int(start)
	iEnd := int(end)

	if iStart < 0 || iStart >= len(str) || iEnd < 0 || iEnd > len(str) || iStart > iEnd {
		return nil, fmt.Errorf("substring() indices out of bounds")
	}

	return str[iStart:iEnd], nil
}

// fnLen returns the length of a string or array.
func (e *Evaluator) fnLen(target interface{}, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("len() takes no arguments, got %d", len(args))
	}

	switch v := target.(type) {
	case string:
		return float64(len(v)), nil
	case []interface{}:
		return float64(len(v)), nil
	case map[string]interface{}:
		return float64(len(v)), nil
	default:
		return nil, fmt.Errorf("len() requires string or collection, got %T", target)
	}
}

// fnRandom returns a random float between 0 and 1.
func (e *Evaluator) fnRandom(target interface{}, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("random() takes no arguments, got %d", len(args))
	}
	return e.context.rand.Float64(), nil
}

// fnRandomInt returns a random integer in range [min, max].
func (e *Evaluator) fnRandomInt(target interface{}, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("random_int() requires 2 arguments, got %d", len(args))
	}

	min, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("random_int() min must be number, got %T", args[0])
	}
	max, ok := toFloat64(args[1])
	if !ok {
		return nil, fmt.Errorf("random_int() max must be number, got %T", args[1])
	}

	iMin := int(min)
	iMax := int(max)

	if iMin > iMax {
		return nil, fmt.Errorf("random_int() min must be <= max")
	}

	return float64(e.context.rand.Intn(iMax-iMin+1) + iMin), nil
}

// fnRandomFloat returns a random float in range [min, max].
func (e *Evaluator) fnRandomFloat(target interface{}, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("random_float() requires 2 arguments, got %d", len(args))
	}

	min, ok := toFloat64(args[0])
	if !ok {
		return nil, fmt.Errorf("random_float() min must be number, got %T", args[0])
	}
	max, ok := toFloat64(args[1])
	if !ok {
		return nil, fmt.Errorf("random_float() max must be number, got %T", args[1])
	}

	if min > max {
		return nil, fmt.Errorf("random_float() min must be <= max")
	}

	return min + e.context.rand.Float64()*(max-min), nil
}

// fnRandomChoice returns a random element from an array.
func (e *Evaluator) fnRandomChoice(target interface{}, args []interface{}) (interface{}, error) {
	arr, ok := target.([]interface{})
	if !ok {
		return nil, fmt.Errorf("random_choice() requires array target, got %T", target)
	}
	if len(args) != 0 {
		return nil, fmt.Errorf("random_choice() takes no arguments, got %d", len(args))
	}
	if len(arr) == 0 {
		return nil, fmt.Errorf("random_choice() requires non-empty array")
	}

	return arr[e.context.rand.Intn(len(arr))], nil
}

// Helper functions

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}
