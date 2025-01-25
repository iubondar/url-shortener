package main

import (
	"fmt"
	"io"
	"net/http"
)

func shortenHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	res.Write([]byte(fmt.Sprintf("Body: %s", body)))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortenHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
