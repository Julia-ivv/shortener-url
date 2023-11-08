package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
)

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileURLs struct {
	fileName string
	Urls     []URL
}

func NewFileURLs(fileName string) (*FileURLs, error) {
	urls := []URL{}
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		url := URL{}
		data := scan.Bytes()
		err := json.Unmarshal(data, &url)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}

	return &FileURLs{
		fileName: fileName,
		Urls:     urls,
	}, nil
}

func (f *FileURLs) GetURL(_ context.Context, shortURL string) (originURL string, ok bool) {
	for _, v := range f.Urls {
		if v.ShortURL == shortURL {
			return v.OriginalURL, true
		}
	}
	return "", false
}

func (f *FileURLs) AddURL(_ context.Context, originURL string) (shortURL string, err error) {
	short := GenerateRandomString(LengthShortURL)
	url := URL{
		ShortURL:    short,
		OriginalURL: originURL,
	}

	f.Urls = append(f.Urls, url)

	file, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	data, err := json.Marshal(&url)
	if err != nil {
		return "", err
	}
	if _, err := wr.Write(data); err != nil {
		return "", err
	}
	if err := wr.WriteByte('\n'); err != nil {
		return "", err
	}

	return short, wr.Flush()
}
