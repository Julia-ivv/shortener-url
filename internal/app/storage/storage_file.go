package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type FileURL struct {
	UserID      int    `json:"user_id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileURLs struct {
	sync.RWMutex
	fileName string
	Urls     []FileURL
}

func NewFileURLs(fileName string) (*FileURLs, error) {
	urls := make([]FileURL, 0)
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scan := bufio.NewScanner(file)
	for scan.Scan() {
		url := FileURL{}
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

func (f *FileURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	// получает урл без учета пользователя
	f.RLock()
	defer f.RUnlock()

	for _, v := range f.Urls {
		if v.ShortURL == shortURL {
			return v.OriginalURL, false, true
		}
	}
	return "", false, false
}

func (f *FileURLs) AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error) {
	short, err := GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}
	url := FileURL{
		UserID:      userID,
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

func (f *FileURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error) {
	var allData []byte
	urls := make([]FileURL, 0)
	for _, v := range originURLBatch {
		sURL, err := GenerateRandomString(LengthShortURL)
		if err != nil {
			return nil, err
		}
		shortURLBatch = append(shortURLBatch, ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      baseURL + sURL,
		})
		url := FileURL{
			UserID:      userID,
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

func (f *FileURLs) GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []UserURL, err error) {
	f.RLock()
	defer f.RUnlock()

	for _, v := range f.Urls {
		if v.UserID == userID {
			userURLs = append(userURLs, UserURL{
				ShortURL:    baseURL + v.ShortURL,
				OriginalURL: v.OriginalURL,
			})
		}
	}
	return userURLs, nil
}

func (f *FileURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
	return nil
}

func (f *FileURLs) PingStor(ctx context.Context) error {
	_, err := os.Stat(f.fileName)
	if os.IsNotExist(err) {
		return errors.New("file not exists")
	}
	return nil
}

func (f *FileURLs) Close() error {
	return nil
}
