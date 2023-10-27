package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/Julia-ivv/shortener-url.git/internal/app/tools"
)

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileWork struct {
	file    *os.File
	writer  *bufio.Writer
	scanner *bufio.Scanner
}

type FileURLs struct {
	file *FileWork
	Urls []URL
}

func NewFileWork(fileName string) (*FileWork, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileWork{
		file:    file,
		writer:  bufio.NewWriter(file),
		scanner: bufio.NewScanner(file),
	}, nil
}

func NewFileURLs(fw *FileWork) (*FileURLs, error) {
	urls := []URL{}
	for fw.scanner.Scan() {
		url := URL{}
		data := fw.scanner.Bytes()
		err := json.Unmarshal(data, &url)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := fw.scanner.Err(); err != nil {
		return nil, err
	}

	return &FileURLs{
		file: fw,
		Urls: urls,
	}, nil
}

func (f *FileURLs) GetURL(shortURL string) (originURL string, ok bool) {
	for _, v := range f.Urls {
		if v.ShortURL == shortURL {
			return v.OriginalURL, true
		}
	}
	return "", false
}

func (f *FileURLs) AddURL(originURL string) (shortURL string, err error) {
	short := tools.GenerateRandomString(tools.LengthShortURL)
	url := URL{
		ShortURL:    short,
		OriginalURL: originURL,
	}

	f.Urls = append(f.Urls, url)

	data, err := json.Marshal(&url)
	if err != nil {
		return "", err
	}
	if _, err := f.file.writer.Write(data); err != nil {
		return "", err
	}
	if err := f.file.writer.WriteByte('\n'); err != nil {
		return "", err
	}

	return short, f.file.writer.Flush()
}

func (f *FileURLs) Close() error {
	return f.file.file.Close()
}
