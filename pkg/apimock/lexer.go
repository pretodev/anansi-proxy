package apimock

import (
	"regexp"
	"strconv"
	"strings"
)

// TokenType represents the type of a lexical token in an .apimock file
type TokenType int

const (
	TokenBlankLine TokenType = iota
	TokenRequestLine
	TokenPathContinuation
	TokenQueryParam
	TokenHeader
	TokenResponseStart
	TokenBodyLine
)

// Token produced by the Lexer
type Token struct {
	Type          TokenType
	Line          int
	Raw           string

	// Request line
	Method        string
	Path          string
	PathSegments  []PathSegment

	// Path continuation
	PathContinuation string

	// Query param
	Key           string
	Value         string

	// Response start
	StatusCode    int
	Description   string
}

// Regular expressions used by the lexer (extracted from parser)
var (
	httpMethodRegex          = regexp.MustCompile(`^(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS|TRACE|CONNECT)\s`)
	pathStartRegex           = regexp.MustCompile(`^(/[a-zA-Z0-9_.\-{}]+)+`)
	pathSegmentRegex         = regexp.MustCompile(`/([a-zA-Z0-9_.\-]+|\{[a-zA-Z0-9_.\-]+\})`)
	pathContRegex            = regexp.MustCompile(`^\s+(/[a-zA-Z0-9_.\-{}]+|\?[a-zA-Z0-9_.\-]+=\S+|&[a-zA-Z0-9_.\-]+=\S+)`)
	queryParamRegex          = regexp.MustCompile(`([a-zA-Z0-9_.\-]+)=(\S+)`)
	responseLineCaptureRegex = regexp.MustCompile(`^--\s*(\d{3}):\s*(.*)`)
	propertyCaptureRegex     = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_.\-]*):\s*(.+)`)
)

// Lexer reads .apimock lines and produces tokens
type Lexer struct {
	lines []string
}

func NewLexer(lines []string) *Lexer {
	return &Lexer{lines: lines}
}

// Lex converts the file contents to a list of tokens
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

// parsePathSegments parses a path into PathSegment tokens
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