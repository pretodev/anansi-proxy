package endpoint

import (
	"fmt"
	"net/http"
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

func EmptyResponse() Response {
	return Response{
		StatusCode: http.StatusNoContent,
	}
}

type EndpointSchema struct {
	Route     string
	Accept    string
	Body      string
	Validator SchemaValidator
	Responses map[int][]Response
}

func (e *EndpointSchema) SliceResponses() []Response {
	var responses []Response
	for _, respList := range e.Responses {
		responses = append(responses, respList...)
	}
	return responses
}

func (e *EndpointSchema) CountResponses() int {
	return len(e.SliceResponses())
}

func (e *EndpointSchema) GetResponseByStatusCode(statusCode int) (Response, bool) {
	if responses, ok := e.Responses[statusCode]; ok && len(responses) > 0 {
		return responses[0], true
	}
	return Response{}, false
}

func (e *EndpointSchema) String() string {
	output := fmt.Sprintf("Route: %s\n", e.Route)
	output += fmt.Sprintf("%s: %s\n", RequestAcceptPropertyName, e.Accept)
	if e.Body != "" {
		output += fmt.Sprintf("Schema: %s\n", e.Body)
	}
	// output += "Responses: \n"
	// for i, r := range e.Responses {
	// 	output += "--------------\n"
	// 	output += fmt.Sprintf("[%d] %d %s\n", i, r.StatusCode, r.Title)
	// 	output += fmt.Sprintf("Body: %s\n", r.Body)
	// 	output += fmt.Sprintf("%s: %s\n", ResponseContentTypePropertyName, r.ContentType)
	// }
	return output
}
