package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Response struct {
	Title       string
	Body        string
	ContentType string
	StatusCode  int
}

func Parse(filePath string) ([]Response, error) {
	content, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("file '%s' is empty", filePath)
	}

	blocks := strings.SplitSeq(content, "###")

	var responses []Response
	for block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		res, err := parseBlock(block)
		if err != nil {
			return nil, fmt.Errorf("failed to parse block: %w", err)
		}
		responses = append(responses, res)
	}

	if len(responses) == 0 {
		return nil, fmt.Errorf("no valid responses found in file '%s'", filePath)
	}

	return responses, nil
}

func parseBlock(block string) (Response, error) {
	lines := strings.Split(block, "\n")

	response := Response{
		Title:       strings.TrimSpace(lines[0]),
		StatusCode:  200,
		ContentType: "text/plain",
	}

	var bodyLines []string
	parsingBody := false

	for i := 1; i < len(lines); i++ {
		line := lines[i]

		if parsingBody {
			bodyLines = append(bodyLines, line)
			continue
		}

		if strings.TrimSpace(line) == "" {
			parsingBody = true
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			if err := parseHeader(&response, parts); err != nil {
				return response, err
			}
			continue
		}

		parsingBody = true
		bodyLines = append(bodyLines, line)
	}

	response.Body = strings.Join(bodyLines, "\n")

	return response, nil
}

func parseHeader(res *Response, parts []string) error {
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch strings.ToLower(key) {
	case "status-code":
		code, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid status code '%s': %w", value, err)
		}
		if code < 100 || code > 599 {
			return fmt.Errorf("status code %d is out of valid range (100-599)", code)
		}
		res.StatusCode = code
	case "content-type":
		if value == "" {
			return fmt.Errorf("content-type cannot be empty")
		}
		res.ContentType = value
	}

	return nil
}

func readFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	defer file.Close()

	content := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}
	return content, nil
}
