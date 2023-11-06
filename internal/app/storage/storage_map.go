package storage

type MapURLs struct {
	originalURLs map[string]string
}

func (urls *MapURLs) GetURL(shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	originURL, ok = urls.originalURLs[shortURL]
	return originURL, ok
}

func (urls *MapURLs) AddURL(originURL string) (shortURL string, err error) {
	// добавить новый урл
	short := GenerateRandomString(LengthShortURL)
	urls.originalURLs[short] = originURL
	return short, nil
}
