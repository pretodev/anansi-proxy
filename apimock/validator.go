package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Token types
type TokenType int

const (
	TokenHTTPMethod TokenType = iota
	TokenPath
	TokenPathSegment
	TokenQueryParam
	TokenProperty
	TokenStatusCode
	TokenDescription
	TokenBody
	TokenBlankLine
	TokenEOF
	TokenError
)

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Content string
}

// Validator validates .apimock files against the EBNF grammar
type Validator struct {
	filename string
	lines    []string
	lineNum  int
	errors   []string
}

// NewValidator creates a new validator for a file
func NewValidator(filename string) (*Validator, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	return &Validator{
		filename: filename,
		lines:    lines,
		lineNum:  0,
		errors:   make([]string, 0),
	}, nil
}

// Regular expressions based on EBNF grammar
var (
	httpMethodRegex   = regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE|CONNECT)\s`)
	pathStartRegex    = regexp.MustCompile(`^(/[a-zA-Z0-9_.\-{}]+)+`)
	pathContRegex     = regexp.MustCompile(`^\s+(/[a-zA-Z0-9_.\-{}]+|\?[a-zA-Z0-9_.\-]+=\S+|&[a-zA-Z0-9_.\-]+=\S+)`)
	propertyRegex     = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.\-]*:\s*.+`)
	responseLineRegex = regexp.MustCompile(`^--\s*\d{3}:\s*.*`)
	statusCodeRegex   = regexp.MustCompile(`\d{3}`)
	identifierRegex   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.\-]*$`)
)

// Validate validates the entire file
func (v *Validator) Validate() bool {
	fmt.Printf("🔍 Validating: %s\n", filepath.Base(v.filename))

	// apimock = [ request_section , blank_lines ] , response_section , { blank_lines , response_section }

	// Skip leading blank lines
	v.skipBlankLines()

	// Optional request section
	if v.lineNum < len(v.lines) && !v.isResponseLine(v.currentLine()) {
		if !v.validateRequestSection() {
			if len(v.errors) == 0 {
				v.addError("Request section validation failed")
			}
			// Don't return yet - let it report errors
		} else {
			v.skipBlankLines()
		}
	}

	// At least one response section
	if v.lineNum >= len(v.lines) || !v.validateResponseSection() {
		if len(v.errors) == 0 {
			v.addError("Expected at least one response section starting with '--'")
		}
		// Continue to error reporting
	} else {
		// Additional response sections
		for v.lineNum < len(v.lines) {
			v.skipBlankLines()
			if v.lineNum >= len(v.lines) {
				break
			}
			if !v.validateResponseSection() {
				break
			}
		}
	}

	// Report results
	if len(v.errors) == 0 {
		fmt.Printf("  ✅ Valid\n\n")
		return true
	}

	fmt.Printf("  ❌ Invalid - %d error(s):\n", len(v.errors))
	for _, err := range v.errors {
		fmt.Printf("     • %s\n", err)
	}
	fmt.Println()
	return false
}

// validateRequestSection validates the request section
func (v *Validator) validateRequestSection() bool {
	// request_section = method_line , { property } , [ blank_line , body ]

	if !v.validateMethodLine() {
		return false
	}

	// Properties (headers)
	for v.lineNum < len(v.lines) {
		line := v.currentLine()
		if v.isBlankLine(line) || v.isResponseLine(line) {
			break
		}
		if propertyRegex.MatchString(line) {
			v.lineNum++
		} else {
			// Not a property - must be start of body
			break
		}
	}

	// Optional body (may have blank line before it)
	if v.lineNum < len(v.lines) && v.isBlankLine(v.currentLine()) {
		v.lineNum++ // skip blank line
	}

	// Read body lines until response section or end
	for v.lineNum < len(v.lines) {
		line := v.currentLine()
		// Stop if we hit a response line
		if v.isResponseLine(line) {
			break
		}
		// Stop if we hit blank lines followed by a response
		if v.isBlankLine(line) {
			// Look ahead to see if response is coming
			nextNonBlank := v.lineNum + 1
			for nextNonBlank < len(v.lines) && v.isBlankLine(v.lines[nextNonBlank]) {
				nextNonBlank++
			}
			if nextNonBlank < len(v.lines) && v.isResponseLine(v.lines[nextNonBlank]) {
				break
			}
		}
		v.lineNum++
	}

	return true
}

// validateMethodLine validates HTTP method and path
func (v *Validator) validateMethodLine() bool {
	// method_line = [ http_method , SP ] , path_start , EOL , { path_continuation }

	if v.lineNum >= len(v.lines) {
		v.addError("Expected method line or path")
		return false
	}

	line := v.currentLine()

	// Check for HTTP method (optional)
	hasMethod := httpMethodRegex.MatchString(line)

	// Must have path
	if !pathStartRegex.MatchString(line) && !hasMethod {
		v.addError(fmt.Sprintf("line %d: Invalid method line format: '%s'", v.lineNum+1, line))
		return false
	}

	v.lineNum++

	// Path continuations (indented lines)
	for v.lineNum < len(v.lines) {
		line := v.currentLine()
		if pathContRegex.MatchString(line) {
			v.lineNum++
		} else {
			break
		}
	}

	return true
}

// validateResponseSection validates a response section
func (v *Validator) validateResponseSection() bool {
	// response_section = response_line , { property } , [ blank_line , response_body ]

	if v.lineNum >= len(v.lines) {
		return false
	}

	// Response line: -- 200: Description
	line := v.currentLine()
	if !responseLineRegex.MatchString(line) {
		v.addError(fmt.Sprintf("line %d: Invalid response line format: '%s'", v.lineNum+1, line))
		return false
	}

	// Validate status code
	matches := statusCodeRegex.FindString(line)
	if matches == "" {
		v.addError(fmt.Sprintf("line %d: Missing or invalid status code", v.lineNum+1))
		return false
	}

	v.lineNum++

	// Properties (headers)
	for v.lineNum < len(v.lines) {
		line := v.currentLine()
		if v.isBlankLine(line) {
			break
		}
		if v.isResponseLine(line) {
			// Next response section
			return true
		}
		if propertyRegex.MatchString(line) {
			v.lineNum++
		} else {
			v.addError(fmt.Sprintf("line %d: Invalid response property: '%s'", v.lineNum+1, line))
			return false
		}
	}

	// Optional response body (after blank line)
	if v.lineNum < len(v.lines) && v.isBlankLine(v.currentLine()) {
		v.lineNum++ // skip blank line
		// Read body lines until next response or EOF
		for v.lineNum < len(v.lines) {
			line := v.currentLine()
			if v.isResponseLine(line) {
				break
			}
			v.lineNum++
		}
	}

	return true
}

// Helper methods

func (v *Validator) currentLine() string {
	if v.lineNum >= len(v.lines) {
		return ""
	}
	return v.lines[v.lineNum]
}

func (v *Validator) peekLine(offset int) string {
	idx := v.lineNum + offset
	if idx >= len(v.lines) {
		return ""
	}
	return v.lines[idx]
}

func (v *Validator) isBlankLine(line string) bool {
	return strings.TrimSpace(line) == ""
}

func (v *Validator) isResponseLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "--")
}

func (v *Validator) isBodyLine(line string) bool {
	// Body can be anything that's not a property or response line
	trimmed := strings.TrimSpace(line)
	return trimmed != "" && !propertyRegex.MatchString(line) && !v.isResponseLine(line)
}

func (v *Validator) skipBlankLines() {
	for v.lineNum < len(v.lines) && v.isBlankLine(v.currentLine()) {
		v.lineNum++
	}
}

func (v *Validator) addError(msg string) {
	v.errors = append(v.errors, msg)
}

// ValidateDirectory validates all .apimock files in a directory
func ValidateDirectory(dir string) (int, int, error) {
	pattern := filepath.Join(dir, "*.apimock")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return 0, 0, err
	}

	if len(files) == 0 {
		return 0, 0, fmt.Errorf("no .apimock files found in %s", dir)
	}

	valid := 0
	invalid := 0

	for _, file := range files {
		validator, err := NewValidator(file)
		if err != nil {
			fmt.Printf("❌ Error reading %s: %v\n", filepath.Base(file), err)
			invalid++
			continue
		}

		if validator.Validate() {
			valid++
		} else {
			invalid++
		}
	}

	return valid, invalid, nil
}
