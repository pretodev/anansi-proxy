package parser

import (
	"bufio"
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
		return nil, err
	}

	blocks := strings.SplitSeq(content, "###")

	var responses []Response
	for block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		res, err := parseBlock(block)
		if err != nil {
			return nil, err
		}
		responses = append(responses, res)
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
			return err
		}
		res.StatusCode = code
	case "content-type":
		res.ContentType = value
	}

	return nil
}

func readFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	content := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return content, nil
}
