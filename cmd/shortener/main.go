package main

import (
	"io"
	"net/http"

	"github.com/iubondar/url-shortener/internal/app/strings"
)

const idLength int = 8

var urlsToIds map[string]string
var idsToURLs map[string]string

func createIdHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), 500)
		return
	}

	url := string(body)
	id, ok := urlsToIds[url]
	if !ok {
		id = strings.RandString(idLength)
		urlsToIds[url] = id
		idsToURLs[id] = url
		res.WriteHeader(http.StatusCreated)
	}

	res.Write([]byte("http://" + req.Host + "/" + id))
}

func retrieveURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	id := req.PathValue("id")
	if len(id) == 0 {
		http.Error(res, "Can't find id parameter in query path", http.StatusBadRequest)
		return
	}

	url, ok := idsToURLs[id]
	if !ok {
		http.Error(res, "Can't find URL by id: "+id, http.StatusBadRequest)
		return
	}

	res.Header().Add("Location", url)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	urlsToIds = make(map[string]string)
	idsToURLs = make(map[string]string)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, createIdHandler)
	mux.HandleFunc(`/{id}`, retrieveURLHandler)

	err := http.ListenAndServe(`localhost:8080`, mux)
	if err != nil {
		panic(err)
	}
}
