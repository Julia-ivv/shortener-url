package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

var urls map[string]string
var inc int = 100

func main() {
	urls = make(map[string]string)
	urls["EwHXdJfB"] = "https://practicum.yandex.ru/"

	mux := http.NewServeMux()

	mux.HandleFunc(`/`, postUrl)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}

func postUrl(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		if req.URL.Path != "/" {
			http.Error(res, "400 bad request", http.StatusBadRequest)
			return
		}
		postUrl, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		inc++
		shortURL := strconv.Itoa(inc)
		urls[shortURL] = string(postUrl)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte("http://" + req.Host + "/" + shortURL))
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodGet:
		parseURL := strings.Split(req.URL.Path, "/")
		shortURL := parseURL[len(parseURL)-1]
		originURL, ok := urls[shortURL]
		if !ok {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originURL)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}

/*
Эндпоинт с методом `POST` и путём `/`. Сервер принимает в теле запроса строку URL как `text/plain` и
возвращает ответ с кодом `201` и сокращённым URL как `text/plain`.
    Пример запроса к серверу:
    POST / HTTP/1.1
    Host: localhost:8080
    Content-Type: text/plain
    https://practicum.yandex.ru/
    Пример ответа от сервера:
    HTTP/1.1 201 Created
    Content-Type: text/plain
    Content-Length: 30
    http://localhost:8080/EwHXdJfB

Эндпоинт с методом `GET` и путём `/{id}`, где `id` — идентификатор сокращённого URL
	(например, `/EwHXdJfB`).
	В случае успешной обработки запроса сервер возвращает ответ с кодом `307`
	и оригинальным URL в HTTP-заголовке `Location`.
    Пример запроса к серверу:
    GET /EwHXdJfB HTTP/1.1
    Host: localhost:8080
    Content-Type: text/plain
    Пример ответа от сервера:
    HTTP/1.1 307 Temporary Redirect
    Location: https://practicum.yandex.ru/

На любой некорректный запрос сервер должен возвращать ответ с кодом `400`Bad Request.

*/
