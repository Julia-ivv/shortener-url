package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
)

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileURLs struct {
	sync.RWMutex
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
	f.RLock()
	defer f.RUnlock()

	for _, v := range f.Urls {
		if v.ShortURL == shortURL {
			return v.OriginalURL, true
		}
	}
	return "", false
}

func (f *FileURLs) AddURL(_ context.Context, originURL string) (shortURL string, err error) {
	short, err := GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}
	url := URL{
		ShortURL:    short,
		OriginalURL: originURL,
	}

	f.Lock()
	defer f.Unlock()

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

	f.Urls = append(f.Urls, url)

	return short, wr.Flush()
}

func (f *FileURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string) (shortURLBatch []ResponseBatch, err error) {
	var allData []byte
	urls := make([]URL, 0)
	for _, v := range originURLBatch {
		sURL, err := GenerateRandomString(LengthShortURL)
		if err != nil {
			return nil, err
		}
		shortURLBatch = append(shortURLBatch, ResponseBatch{
			CorrelationId: v.CorrelationId,
			ShortURL:      baseURL + sURL,
		})
		url := URL{
			ShortURL:    sURL,
			OriginalURL: v.OriginalURL,
		}
		urls = append(urls, url)
		data, err := json.Marshal(url)
		if err != nil {
			return nil, err
		}
		allData = append(allData, data...)
		allData = append(allData, '\n')
	}

	f.Lock()
	defer f.Unlock()

	file, err := os.OpenFile(f.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	wr := bufio.NewWriter(file)
	if _, err := wr.Write(allData); err != nil {
		return nil, err
	}
	err = wr.Flush()
	if err != nil {
		return nil, err
	}

	f.Urls = append(f.Urls, urls...)

	return shortURLBatch, nil
}
