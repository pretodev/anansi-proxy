package main

import (
	"fmt"

	"github.com/pretodev/anansi-proxy/internal/parser"
)

func main() {
	resPath := "./resources/api-rest.httpr"

	res, err := parser.Parse(resPath)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}
