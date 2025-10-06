# APIMock Conditions Language Specification

## Table of Contents

1. [Overview](#overview)
2. [Core Concepts](#core-concepts)
3. [Syntax](#syntax)
4. [Data Types](#data-types)
5. [Operators](#operators)
6. [Built-in Functions](#built-in-functions)
7. [Context Variables](#context-variables)
8. [Advanced Features](#advanced-features)
9. [Examples](#examples)
10. [Best Practices](#best-practices)

---

## Overview

APIMock Conditions is a domain-specific language (DSL) designed for controlling API mock responses through conditional logic. It provides a simple, declarative way to define when specific responses should be returned based on request data, state, and other conditions.

### Key Features

- **Immutable**: All values are immutable by design
- **Type-safe**: Support for number, boolean, string, and table types
- **Declarative**: Conditions are evaluated top-to-bottom, first match wins
- **Expressive**: Rich set of operators and built-in functions
- **Deterministic**: Random functions use request-based seeding for reproducibility

---

## Core Concepts

### Condition Blocks

A condition block consists of one or more condition lines (starting with `>`) followed by a response:

```apimock
> condition1
> condition2
> condition3

{
  "response": "data"
}
```

### Evaluation Rules

1. **Top-to-Bottom Evaluation**: The server evaluates condition blocks sequentially from top to bottom
2. **First Match Wins**: The first block where ALL conditions evaluate to `True` returns its response
3. **Conjunction by Default**: Multiple `>` lines are combined with AND logic
4. **Disjunction with `or`**: Use `or` keyword to create OR logic between conditions
5. **Empty Condition**: A single `>` with no expression evaluates to `False`

### Truth and Falsy Values

**Truthy Values** (evaluate to `True`):
- `True` (boolean)
- Any non-zero number (e.g., `12`, `23.7`, `-5`)
- Any non-empty string (e.g., `"Vanessa"`)
- Any non-empty table (e.g., `{key = "value"}`, `{1, 2, 3}`)

**Falsy Values** (evaluate to `False`):
- `False` (boolean)
- `0` (number)
- `""` (empty string)
- `{}` (empty table)
- Empty condition (`>`)

---

## Syntax

### Basic Condition Syntax

```apimock
> expression
```

### Attribution (Assignment)

```apimock
> value >> variable_name
```

The attribution operator (`>>`) assigns a value to a variable and evaluates to `True`.

### Comments

Comments use the `#` symbol:

```apimock
> True  # This is an inline comment
```

### Multi-line Conditions

#### Conjunction (AND) - Default Behavior

All conditions must be `True`:

```apimock
> condition1
> condition2
> condition3
```

**Result**: `condition1 AND condition2 AND condition3`

#### Disjunction (OR) - Using `or` Keyword

At least one condition must be `True`:

```apimock
> condition1
> or condition2
> or condition3
```

**Result**: `condition1 OR condition2 OR condition3`

#### Mixed Logic

```apimock
> condition1
> condition2
> or condition3
```

**Result**: `(condition1 AND condition2) OR condition3`

---

## Data Types

### Number

Integer or floating-point numbers:

```apimock
> 42              # Integer
> 3.14            # Float
> -17             # Negative
> 0               # Zero (falsy)
```

### Boolean

Logical values:

```apimock
> True            # Boolean true
> False           # Boolean false
```

### String

Text enclosed in double quotes:

```apimock
> "Hello World"   # String
> ""              # Empty string (falsy)
> "123"           # String of numbers
```

### Table

Tables can be used as arrays or dictionaries:

#### Array-style (indexed):

```apimock
> {1, 2, 3, 4, 5}
> {"apple", "banana", "orange"}
```

#### Dictionary-style (key-value):

```apimock
> {name = "Silas", age = 30}
> {key = "value", active = True}
```

#### Empty table (falsy):

```apimock
> {}
```

### Range

Ranges generate sequences of numbers:

```apimock
> 1..10           # Range from 1 to 10 [1,2,3,4,5,6,7,8,9,10]
> 0..100          # Range from 0 to 100
> 10..20          # Range from 10 to 20
```

---

## Operators

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal to | `x == 5` |
| `!=` | Not equal to | `x != 0` |
| `>` | Greater than | `x > 10` |
| `<` | Less than | `x < 20` |
| `>=` | Greater than or equal | `x >= 5` |
| `<=` | Less than or equal | `x <= 15` |

### Logical Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `not` | Logical NOT | `not False` |
| `and` | Logical AND | `True and False` |
| `or` | Logical OR | `True or False` |

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `5 + 3` |
| `-` | Subtraction | `10 - 4` |
| `*` | Multiplication | `6 * 7` |
| `/` | Division | `20 / 4` |
| `%` | Modulo | `15 % 4` |
| `//`| Integer division | `15 // 4` |

### String Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `..` | Concatenation | `"Hello" .. " World"` |


### Attribution Operator

| Operator | Description | Example |
|----------|-------------|---------|
| `>>` | Assignment | `value >> variable` |

---

## Built-in Functions

All built-in functions are called with a dot prefix (`.`). They operate on the piped value or are called globally.

### String Functions

#### `.split`

Splits a string into a table based on a delimiter.

```apimock
> "2025-10-06" >> date
> date >> .split "-" >> year, month, day
> year == "2025"  # -> True
```

#### `.contains`

Checks if a string/table contains a substring/element.

```apimock
> email >> .contains "@"  # -> True if email contains "@"
> tags >> .contains "invalid"  # -> True if tags doesn't contain "invalid" elemement
```

#### `.not_contains`

Checks if a string or table does **not** contain a substring or element.

```apimock
> email >> .not_contains "@"
>  tags >> .not_contains "urgent"  # True if "urgent" is present in tags
```

#### `.trim`

Removes leading and trailing whitespace.

```apimock
> "  hello  " >> text
> text >> .trim >> trimmed
> trimmed == "hello"  # -> True
```

### Table/Array Functions

#### `.contains`

Checks if a table contains a specific element.

```apimock
> {1, 2, 3, 4, 5} >> numbers
> numbers >> .contains 3  # -> True
```

### Type Checking Functions

#### `.is_string`, `.is_number`, `.is_boolean`, `.is_table`

Check the type of a value.

```apimock
> city >> .is_string >> is_str  # -> True if city is a string
> age >> .is_number >> is_num   # -> True if age is a number
```

### Math Functions

#### `.round`

Rounds a number to the nearest integer.

```apimock
> 8.7 >> grade
> grade >> .round >> rounded  # -> 9
```

#### `.floor`

Returns the largest integer less than or equal to the number.

```apimock
> 8.7 >> value
> value >> .floor >> floored  # -> 8
```

#### `.ceil`

Returns the smallest integer greater than or equal to the number.

```apimock
> 8.2 >> value
> value >> .ceil >> ceiled  # -> 9
```

#### `.abs`

Returns the absolute value of a number.

```apimock
> -5 >> negative
> negative >> .abs >> absolute  # -> 5
```

### Random Functions

All random functions are deterministic based on the request context, ensuring reproducibility for testing.

#### `.random_bool`

Returns a random boolean value.

```apimock
> .random_bool >> is_lucky
```

#### `.random_int`

Returns a random integer within a range.

```apimock
> .random_int 1 10 >> dice  # Random number between 1 and 10
```

#### `.random_float`

Returns a random float within a range.

```apimock
> .random_float 0.0 1.0 >> probability
```

---

## Context Variables

Context variables provide access to request data and mock state without requiring prefixes in most cases.

### Request Context

Access HTTP request information:

#### `method`

The HTTP method of the request.

```apimock
> method == "POST"
> method == "GET"
```

#### `path`

The request path.

```apimock
> path >> .contains "/users/"
```

#### `headers`

Access request headers using dictionary syntax.

```apimock
> headers["Authorization"] >> token
> headers["Content-Type"] >> .contains "json"
```

#### `query`

Access query parameters.

```apimock
> query.page >> page_num
> page_num > 1
```

#### `body`

Access request body fields.

```apimock
> body.name >> username
> body.email >> .contains "@"
> body.age >= 18
```

### Mock State Context

#### `call_count`

The number of times this endpoint has been called.

```apimock
> call_count > 5  # True after 5 calls
```

#### `timestamp`

Current timestamp in ISO 8601 format.

```apimock
> timestamp >> now
> now >> .contains "2025"
```

#### `date`

Current date in YYYY-MM-DD format.

```apimock
> date >> today
> today == "2025-10-06"
```

---

## Advanced Features

### Destructuring

Unpack tables into multiple variables:

```apimock
> "Silas Ribeiro Prado" >> full_name
> full_name >> .split " " >> first, middle, last
> first == "Silas"  # -> True
> last == "Prado"   # -> True
```

Multiple levels of destructuring:

```apimock
> "2025-10-06T14:30:00" >> timestamp
> timestamp >> .split "T" >> date_part, time_part
> date_part >> .split "-" >> year, month, day
> year == "2025"  # -> True
```

### Ranges

Generate sequences and use them with functions:

Check if value is in range:

```apimock
> 1..10 >> valid_range
> valid_range >> .contains 5  # -> True
```

### Table Property Access

Access table properties using dot notation:

```apimock
> {name = "Silas", city = "Salvador"} >> person
> person.name == "Silas"  # -> True
> person.city == "Salvador"  # -> True
```

Nested access:

```apimock
> body.address.city == "Salvador"
> body.user.profile.age > 18
```

### Chaining Operations

Chain multiple operations using the pipe operator:

```apimock
> email >> .split "@" >> local, domain
> domain >> .contains "gmail"
```

```apimock
> "  hello  " >> text
> text >> .trim >> .contains "hello"  # -> True
```

---

## Examples

### Example 1: Rate Limiting

```apimock
-- 429: Too Many Requests
ContentType: application/json
> call_count > 5

{
  "error": "Too Many Requests",
  "code": 429,
  "message": "Rate limit exceeded. Try again later."
}

-- 200: Success
ContentType: application/json
> call_count <= 5

{
  "status": "success"
}
```

### Example 2: Authentication

```apimock
-- 401: Unauthorized - Missing token
ContentType: application/json
> headers["Authorization"] >> token
> token == "" or not token

{
  "error": "Unauthorized",
  "code": 401,
  "message": "Missing authentication token"
}

-- 401: Unauthorized - Invalid token
ContentType: application/json
> headers["Authorization"] >> auth_header
> auth_header >> .split " " >> bearer, token_value
> token_value != "valid-secret-123"

{
  "error": "Unauthorized",
  "code": 401,
  "message": "Invalid token"
}

-- 200: Success
ContentType: application/json
> headers["Authorization"] == "Bearer valid-secret-123"

{
  "status": "authenticated",
  "user": "john_doe"
}
```

### Example 3: Email Validation

```apimock
-- 400: Bad Request - Invalid email
ContentType: application/json
> body.email >> user_email
> user_email >> .not_contains "@"
> or user_email >> .not_contains "."

{
  "error": "Bad Request",
  "code": 400,
  "message": "Invalid email format"
}

-- 200: Success
ContentType: application/json
> body.email >> .contains "@"

{
  "status": "success",
  "email": "{{body.email}}"
}
```

### Example 4: Age Validation

```apimock
-- 400: Bad Request - Underage
ContentType: application/json
> body.birthdate >> birthdate
> .date >> current_date
> birthdate >> .split "-" >> birth_y, birth_m, birth_d
> current_date >> .split "-" >> curr_y, curr_m, curr_d
> curr_y - birth_y >> age
> age < 18

{
  "error": "Bad Request",
  "code": 400,
  "message": "User must be at least 18 years old",
  "calculatedAge": "{{age}}"
}

-- 201: Created
ContentType: application/json
> body.birthdate >> birthdate
> .date >> current_date
> birthdate >> .split "-" >> birth_y, birth_m, birth_d
> current_date >> .split "-" >> curr_y, curr_m, curr_d
> curr_y - birth_y >= 18

{
  "status": "success",
  "userId": 123
}
```

### Example 5: Random Failures (Chaos Testing)

```apimock
-- 503: Service Unavailable (2% random failure)
ContentType: application/json
> .random_int 100 >> random_num
> random_num <= 2

{
  "error": "Service Unavailable",
  "code": 503,
  "message": "Temporary service failure",
  "retryAfter": 30
}

-- 200: Success
ContentType: application/json
> True

{
  "status": "success"
}
```

### Example 6: Geographic Routing

```apimock
-- 200: South America Region
ContentType: application/json
> headers["X-Country-Code"] >> country
> {"BR", "AR", "CL", "UY", "PY"} >> south_america
> south_america >> .contains country

{
  "region": "South America",
  "country": "{{country}}",
  "server": "sa-east-1"
}

-- 200: North America Region
ContentType: application/json
> headers["X-Country-Code"] >> country
> {"US", "CA", "MX"} >> north_america
> north_america >> .contains country

{
  "region": "North America",
  "country": "{{country}}",
  "server": "us-east-1"
}

-- 200: Default Region
ContentType: application/json
> True

{
  "region": "Europe",
  "server": "eu-west-1"
}
```

### Example 7: Missing Required Fields

```apimock
-- 400: Bad Request - Missing fields
ContentType: application/json
> not body.name
> or not body.email
> or not body.password

{
  "error": "Bad Request",
  "code": 400,
  "message": "Missing required fields: name, email, and password are required"
}

-- 201: Created
ContentType: application/json
> body.name
> body.email
> body.password

{
  "status": "success",
  "userId": 456
}
```

### Example 8: Progressive Retry Behavior

```apimock
-- 500: Internal Server Error (first 2 attempts)
ContentType: application/json
> call_count <= 2

{
  "error": "Internal Server Error",
  "code": 500,
  "message": "Service temporarily unavailable",
  "attempt": "{{call_count}}"
}

-- 200: Success (3rd attempt onwards)
ContentType: application/json
> call_count > 2

{
  "status": "success",
  "message": "Service recovered after retries",
  "attempt": "{{call_count}}"
}
```

---

## Best Practices

### 1. Order Conditions by Specificity

Place more specific conditions before general ones:

```apimock
-- 400: Specific validation error
> body.email == ""

{...}

-- 400: General validation error
> not body.email

{...}
```

### 2. Use Descriptive Variable Names

Make your conditions self-documenting:

```apimock
> headers["Authorization"] >> auth_token
> auth_token >> .split " " >> bearer, token_value
> token_value == "secret-123"
```

### 3. Group Related Conditions

Keep related validation logic together:

```apimock
# Age validation group
> body.birthdate >> birthdate
> .date >> current_date
> birthdate >> .split "-" >> birth_y, birth_m, birth_d
> current_date >> .split "-" >> curr_y, curr_m, curr_d
> curr_y - birth_y >= 18
```

### 4. Use Comments for Clarity

Document complex conditions:

```apimock
# Check if user is authenticated AND has admin role
> headers["Authorization"] >> token
> token >> .contains "Bearer"
> body.role == "admin"
```

### 5. Leverage Immutability

Remember all values are immutable, so you can safely reuse variables:

```apimock
> body.email >> user_email
> user_email >> .contains "@"
> user_email >> .split "@" >> local, domain
> not local == ""
```

### 6. Test Edge Cases

Always include conditions for edge cases:

```apimock
-- Handle empty strings
> not body.name or body.name == ""

-- Handle zero values
> body.age == 0

-- Handle empty tables
> body.tags == {}
```

### 7. Use Ranges for Random Behavior

Use ranges with `.random_int` for controlled randomness:

```apimock
> .random_int 100 >> chance
> chance <= 10  # 10% probability
```

### 8. Keep Conditions Simple

Prefer multiple simple conditions over complex nested logic:

**Good:**
```apimock
> body.email >> .contains "@"
> body.email >> .contains "."
```

**Avoid:**
```apimock
> (body.email >> .contains "@") and (body.email >> .contains ".")
```

### 9. Use Fallback Responses

Always include a default response at the end:

```apimock
-- Specific cases
...

-- Default fallback
> True
{
  "status": "success"
}
```

### 10. Document Response Status Codes

Use clear descriptions in response headers:

```apimock
-- 400: Bad Request - Invalid email format
-- 401: Unauthorized - Missing token
-- 429: Too Many Requests
-- 200: Success
```

---

## Language Grammar Summary

### Tokens

```
COMMENT      := '#' .* '\n'
CONDITION    := '>'
ATTRIBUTION  := '>>'
OR_KEYWORD   := 'or'
AND_KEYWORD  := 'and'
NOT_KEYWORD  := 'not'

NUMBER       := [0-9]+ ('.' [0-9]+)?
BOOLEAN      := 'True' | 'False'
STRING       := '"' .* '"'
IDENTIFIER   := [a-zA-Z_][a-zA-Z0-9_]*

RANGE        := NUMBER '..' NUMBER
TABLE        := '{' (key '=' value | value) (',' (key '=' value | value))* '}'
```

### Operators

```
COMPARISON   := '==' | '!=' | '>' | '<' | '>=' | '<='
ARITHMETIC   := '+' | '-' | '*' | '/' | '%'
LOGICAL      := 'and' | 'or' | 'not'
STRING_OP    := '..'
```

### Built-in Functions

```
BUILTIN      := '.' IDENTIFIER
```

### Syntax Rules

```
condition_block := (condition_line)+ response
condition_line  := '>' expression? ('\n' | EOF)
                 | '>' 'or' expression ('\n' | EOF)
expression      := term ((AND | OR) term)*
term            := factor (comparison_op factor)*
factor          := value | unary_op factor | '(' expression ')'
value           := NUMBER | BOOLEAN | STRING | TABLE | RANGE | IDENTIFIER
                 | value '>>' IDENTIFIER
                 | value BUILTIN
                 | BUILTIN
```

---

## Conclusion

APIMock Conditions provides a powerful yet simple language for controlling mock API responses. Its immutable design, rich set of operators, and built-in functions make it ideal for creating realistic, testable API mocks with complex conditional logic.

For more examples, see the `examples/` directory:
- `logical.apimock` - Comprehensive language reference
- `conditions.apimock` - Real-world condition examples

---

**Version:** 1.0  
**Last Updated:** October 6, 2025
