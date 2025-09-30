package main

import (
	"github.com/pretodev/anansi-proxy/internal/parser"
	"github.com/pretodev/anansi-proxy/internal/ui"
)

func main() {
	resPath := "./resources/api-rest.httpr"

	res, err := parser.Parse(resPath)
	if err != nil {
		panic(err)
	}

	if err := ui.Render(res); err != nil {
		panic(err)
	}
}
