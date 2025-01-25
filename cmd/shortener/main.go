package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/strings"
)

const shortURLLength int = 8

var urls map[string]string

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

	short, ok := urls[string(body)]
	if !ok {
		short = strings.RandString(shortURLLength)
		urls[string(body)] = short
		res.WriteHeader(http.StatusCreated)
	}

	res.Write([]byte(fmt.Sprintf("http://%s/%s", req.Host, short)))
}

func main() {
	urls = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortenHandler)

	err := http.ListenAndServe(`localhost:8080`, mux)
	if err != nil {
		panic(err)
	}
}
