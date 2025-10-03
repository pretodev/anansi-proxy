package endpoint

import (
	"fmt"
)

const (
	ContentTypeHeader               = "Content-Type"
	DefaultContentType              = "text/plain; charset=utf-8"
	RequestAcceptPropertyName       = "Accept"
	ResponseContentTypePropertyName = "ContentType"
)

type Response struct {
	Title       string
	Body        string
	ContentType string
	StatusCode  int
}

type EndpointSchema struct {
	Route     string
	Accept    string
	Body      string
	Responses []Response
}

func (e *EndpointSchema) String() string {
	output := fmt.Sprintf("Route: %s\n", e.Route)
	output += fmt.Sprintf("%s: %s\n", RequestAcceptPropertyName, e.Accept)
	if e.Body != "" {
		output += fmt.Sprintf("Schema: %s\n", e.Body)
	}
	output += "Responses: \n"
	for i, r := range e.Responses {
		output += "--------------\n"
		output += fmt.Sprintf("[%d] %d %s\n", i, r.StatusCode, r.Title)
		output += fmt.Sprintf("Body: %s\n", r.Body)
		output += fmt.Sprintf("%s: %s\n", ResponseContentTypePropertyName, r.ContentType)
	}
	return output
}
