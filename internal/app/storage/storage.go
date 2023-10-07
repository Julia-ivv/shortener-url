package storage

import "strconv"

var inc int

type URLs struct {
	originalURLs map[string]string
}

type Repositories interface {
	GetURL(shortURL string) (originURL string, ok bool)
	AddURL(originURL string) (shortURL string)
}

var Repo URLs

func init() {
	inc = 100
	Repo.originalURLs = make(map[string]string)
	Repo.originalURLs["EwHXdJfB"] = "https://practicum.yandex.ru/"
}

func (urls *URLs) GetURL(shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	originURL, ok = urls.originalURLs[shortURL]
	return originURL, ok
}

func (urls *URLs) AddURL(originURL string) (shortURL string) {
	// добавить новый урл
	inc++
	short := strconv.Itoa(inc)
	urls.originalURLs[short] = originURL
	return short
}
