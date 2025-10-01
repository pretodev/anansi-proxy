package endpoint

const (
	ContentTypeHeader  = "Content-Type"
	DefaultContentType = "text/plain; charset=utf-8"
)

type EndpointSchema struct {
	Route       string
	ContentType string
	Body        string
	Responses   []Response
}

type Response struct {
	Title       string
	Body        string
	ContentType string
	StatusCode  int
}
