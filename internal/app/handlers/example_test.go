package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
)

func ExampleHandlers_PostURL() {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	const userID = 123
	const userContextKey = "user"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post("/", AddContext(hs.PostURL))
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, _ := http.NewRequestWithContext(context.WithValue(context.Background(), userContextKey, userID),
		http.MethodPost, ts.URL+"/", strings.NewReader("https://mail.ru/"))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandlers_PostJSON() {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	path := "/api/shorten"
	const userID = 123
	const userContextKey = "user"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post(path, AddContext(hs.PostJSON))
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, _ := http.NewRequestWithContext(context.WithValue(context.Background(), userContextKey, userID),
		http.MethodPost, ts.URL+path, strings.NewReader(`{"url":"https://mail.ru"}`))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandlers_PostBatch() {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	path := "/api/shorten/batch"
	const userID = 123
	const userContextKey = "user"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post(path, AddContext(hs.PostBatch))
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, _ := http.NewRequestWithContext(context.WithValue(context.Background(), userContextKey, userID),
		http.MethodPost, ts.URL+path, strings.NewReader(`[
			{
				"correlation_id": "ind1",
				"original_url": "https://pract.ru/url1"
			},
			{
				"correlation_id": "ind2",
				"original_url": "https://pract.ru/url2"
			},
			{
				"correlation_id": "ind3",
				"original_url": "https://pract.ru/url3"
			},
			{
				"correlation_id": "ind4",
				"original_url": "https://pract.ru/url4"
			}
		]`))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)

	// Output:
	// 201
}

func ExampleHandlers_GetURL() {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:    testUserID,
		shortURL:  "EwH",
		originURL: "https://practicum.yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}
	path := "/"
	const userID = 123
	const userContextKey = "user"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Get(path+"{shortURL}", AddContext(hs.GetURL))
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, _ := http.NewRequestWithContext(context.WithValue(context.Background(), userContextKey, userID),
		http.MethodGet, ts.URL+path, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Location())

	// Output:
	// 307
	// https://practicum.yandex.ru/
}

func ExampleHandlers_GetUserURLs() {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:    testUserID,
		shortURL:  "EwH",
		originURL: "https://practicum.yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}
	path := "/api/user/urls"
	const userID = 123
	const userContextKey = "user"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Get(path, AddContext(hs.GetUserURLs))
	ts := httptest.NewServer(router)
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, _ := http.NewRequestWithContext(context.WithValue(context.Background(), userContextKey, userID),
		http.MethodGet, ts.URL+path, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(string(body))

	// Output:
	// 200
	// [{"short_url": "http://localhost:8080/EwH", "original_url": "https://practicum.yandex.ru/"}]
}
