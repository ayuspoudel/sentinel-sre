package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

/*
export TARGET_URL=http://localhost:8000/health
go run examples/curl/main.go
>>>> ok
*/
func main() {
	targetUrl := os.Getenv("TARGET_URL")
	resp, err := http.Get(targetUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
}
