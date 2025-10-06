package apimock

import (
	"regexp"
	"strconv"
	"strings"
)

// TokenType represents the type of a lexical token in an .apimock file.
// Each token type corresponds to a different syntactic element in the file format.
type TokenType int

const (
	// TokenBlankLine represents an empty or whitespace-only line
	TokenBlankLine TokenType = iota
	// TokenRequestLine represents the initial request line (method and/or path)
	TokenRequestLine
	// TokenPathContinuation represents a continuation of the path on a new line
	TokenPathContinuation
	// TokenQueryParam represents a query parameter (key=value)
	TokenQueryParam
	// TokenHeader represents an HTTP header (key: value)
	TokenHeader
	// TokenResponseStart represents the start of a response section (-- code: description)
	TokenResponseStart
	// TokenConditionLine represents a condition line (starts with >)
	TokenConditionLine
	// TokenBodyLine represents a line of body content (request or response)
	TokenBodyLine
)

// Token represents a lexical token produced by the Lexer.
// Different fields are populated based on the TokenType.
type Token struct {
	Type TokenType
	Line int
	Raw  string

	// Request line
	Method       string
	Path         string
	PathSegments []PathSegment

	// Path continuation
	PathContinuation string

	// Query param
	Key   string
	Value string

	// Response start
	StatusCode  int
	Description string

	// Condition line
	ConditionExpression string // The full condition expression after >
	IsOrCondition       bool   // true if line starts with "> or"
}

// Regular expressions used by the lexer for pattern matching
var (
	// httpMethodRegex matches HTTP method verbs at the start of a line
	httpMethodRegex = regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE|CONNECT)\s`)
	// pathStartRegex matches URL paths starting with /
	pathStartRegex = regexp.MustCompile(`^(/[a-zA-Z0-9_.\-{}]+)+`)
	// pathSegmentRegex matches individual path segments including parameters
	pathSegmentRegex = regexp.MustCompile(`/([a-zA-Z0-9_.\-]+|\{[a-zA-Z0-9_.\-]+\})`)
	// pathContRegex matches path continuations on indented lines
	pathContRegex = regexp.MustCompile(`^\s+(/[a-zA-Z0-9_.\-{}]+|\?[a-zA-Z0-9_.\-]+=\S+|&[a-zA-Z0-9_.\-]+=\S+)`)
	// queryParamRegex extracts key-value pairs from query parameters
	queryParamRegex = regexp.MustCompile(`([a-zA-Z0-9_.\-]+)=(\S+)`)
	// responseLineCaptureRegex matches response start lines (-- 200: Description)
	responseLineCaptureRegex = regexp.MustCompile(`^--\s*(\d{3}):\s*(.*)`)
	// propertyCaptureRegex matches header-like properties (Key: Value)
	propertyCaptureRegex = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_.\-]*):\s*(.+)`)
	// conditionLineRegex matches condition lines (> expression or > or expression)
	conditionLineRegex = regexp.MustCompile(`^>\s*(.*)`)
	// conditionOrRegex checks if condition starts with "or" keyword
	conditionOrRegex = regexp.MustCompile(`^or\s+(.+)`)
)

// Lexer reads .apimock file lines and produces a stream of tokens.
// It performs lexical analysis to identify syntactic elements.
type Lexer struct {
	lines []string
}

// NewLexer creates a new Lexer for the given lines.
// The lines parameter should contain the complete file content split by newlines.
func NewLexer(lines []string) *Lexer {
	return &Lexer{lines: lines}
}

// Lex converts the file contents to a list of tokens.
// It analyzes each line and produces appropriate tokens based on the content.
// Returns an error if the lexical analysis fails.
func (l *Lexer) Lex() ([]Token, error) {
	tokens := make([]Token, 0)

	for i, line := range l.lines {
		trimmed := strings.TrimSpace(line)

		// Blank line
		if trimmed == "" {
			tokens = append(tokens, Token{Type: TokenBlankLine, Line: i + 1, Raw: line})
			continue
		}

		// Response line
		if m := responseLineCaptureRegex.FindStringSubmatch(line); m != nil {
			code, err := strconv.Atoi(m[1])
			if err != nil {
				code = 0
			}
			tokens = append(tokens, Token{
				Type:        TokenResponseStart,
				Line:        i + 1,
				Raw:         line,
				StatusCode:  code,
				Description: strings.TrimSpace(m[2]),
			})
			continue
		}

		// Condition line (must come before Header check to avoid conflicts)
		if m := conditionLineRegex.FindStringSubmatch(line); m != nil {
			expression := strings.TrimSpace(m[1])

			// Remove inline comments (everything after #)
			if commentIdx := strings.Index(expression, "#"); commentIdx != -1 {
				expression = strings.TrimSpace(expression[:commentIdx])
			}

			// Check if it's an OR condition
			isOrCondition := false
			if orMatch := conditionOrRegex.FindStringSubmatch(expression); orMatch != nil {
				isOrCondition = true
				expression = strings.TrimSpace(orMatch[1])
			}

			tokens = append(tokens, Token{
				Type:                TokenConditionLine,
				Line:                i + 1,
				Raw:                 line,
				ConditionExpression: expression,
				IsOrCondition:       isOrCondition,
			})
			continue
		}

		// Header property
		if m := propertyCaptureRegex.FindStringSubmatch(line); m != nil {
			tokens = append(tokens, Token{Type: TokenHeader, Line: i + 1, Raw: line, Key: m[1], Value: strings.TrimSpace(m[2])})
			continue
		}

		// Path continuation or query params (indented)
		if m := pathContRegex.FindStringSubmatch(line); m != nil {
			cont := strings.TrimSpace(m[1])
			if strings.HasPrefix(cont, "/") {
				segs := parsePathSegments(cont)
				tokens = append(tokens, Token{Type: TokenPathContinuation, Line: i + 1, Raw: line, PathContinuation: cont, PathSegments: segs})
				continue
			}
			if strings.HasPrefix(cont, "?") || strings.HasPrefix(cont, "&") {
				pairs := queryParamRegex.FindAllStringSubmatch(cont[1:], -1)
				for _, pair := range pairs {
					if len(pair) >= 3 {
						tokens = append(tokens, Token{Type: TokenQueryParam, Line: i + 1, Raw: line, Key: pair[1], Value: pair[2]})
					}
				}
				continue
			}
		}

		// Request line: method + path OR path only
		method := ""
		rest := line
		if m := httpMethodRegex.FindStringSubmatch(line); m != nil {
			method = strings.TrimSpace(m[1])
			rest = strings.TrimSpace(line[len(m[0]):])
		}

		path := pathStartRegex.FindString(rest)
		if path == "" {
			path = pathStartRegex.FindString(line)
		}
		if path != "" {
			segs := parsePathSegments(path)
			tokens = append(tokens, Token{Type: TokenRequestLine, Line: i + 1, Raw: line, Method: method, Path: path, PathSegments: segs})
			continue
		}

		// Otherwise, treat as body line
		tokens = append(tokens, Token{Type: TokenBodyLine, Line: i + 1, Raw: line})
	}

	return tokens, nil
}

// parsePathSegments parses a path string into PathSegment tokens.
// It identifies both static segments and parameter placeholders (e.g., {id}).
func parsePathSegments(path string) []PathSegment {
	segments := make([]PathSegment, 0)
	matches := pathSegmentRegex.FindAllStringSubmatch(path, -1)
	for _, match := range matches {
		seg := match[1]
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			name := seg[1 : len(seg)-1]
			segments = append(segments, PathSegment{Value: seg, IsParameter: true, Name: name})
		} else {
			segments = append(segments, PathSegment{Value: seg, IsParameter: false})
		}
	}
	return segments
}
