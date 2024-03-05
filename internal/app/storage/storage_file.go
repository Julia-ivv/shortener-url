package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"sync"
)

// FileURL stores URL information in file.
type FileURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	DeletedFlag bool   `json:"is_deleted"`
	UserID      int    `json:"user_id"`
}

// FileURLs stores information about all URLs in file.
type FileURLs struct {
	fileName string
	file     *os.File
	Urls     []FileURL
	sync.RWMutex
}

// NewFileURLs creates an instance for storing URLs.
func NewFileURLs(fileName string) (*FileURLs, error) {
	urls := make([]FileURL, 0)
	fileRd, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer fileRd.Close()

	scan := bufio.NewScanner(fileRd)
	for scan.Scan() {
		url := FileURL{}
		data := scan.Bytes()
		err = json.Unmarshal(data, &url)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err = scan.Err(); err != nil {
		return nil, err
	}

	fileWr, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileURLs{
		fileName: fileName,
		file:     fileWr,
		Urls:     urls,
	}, nil
}

// GetURL gets the original URL matching the short URL.
func (f *FileURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	f.RLock()
	defer f.RUnlock()

	for _, v := range f.Urls {
		if v.ShortURL == shortURL {
			return v.OriginalURL, v.DeletedFlag, true
		}
	}
	return "", false, false
}

// AddURL adds a new short url.
func (f *FileURLs) AddURL(ctx context.Context, shortURL string, originURL string, userID int) (err error) {
	url := FileURL{
		UserID:      userID,
		ShortURL:    shortURL,
		OriginalURL: originURL,
		DeletedFlag: false,
	}

	f.Lock()
	defer f.Unlock()

	wr := bufio.NewWriter(f.file)
	data, err := json.Marshal(&url)
	if err != nil {
		return err
	}
	if _, err := wr.Write(data); err != nil {
		return err
	}
	if err := wr.WriteByte('\n'); err != nil {
		return err
	}

	f.Urls = append(f.Urls, url)

	return wr.Flush()
}

// AddBatch adds a batch of new short URLs.
func (f *FileURLs) AddBatch(ctx context.Context, shortURLBatch []ResponseBatch, originURLBatch []RequestBatch, userID int) (err error) {
	var allData []byte
	urls := make([]FileURL, 0)
	for k, v := range shortURLBatch {
		url := FileURL{
			UserID:      userID,
			ShortURL:    v.ShortURL,
			OriginalURL: originURLBatch[k].OriginalURL,
			DeletedFlag: false,
		}
		urls = append(urls, url)
		var data []byte
		data, err = json.Marshal(url)
		if err != nil {
			return err
		}
		allData = append(allData, data...)
		allData = append(allData, '\n')
	}

	f.Lock()
	defer f.Unlock()

	wr := bufio.NewWriter(f.file)
	if _, err = wr.Write(allData); err != nil {
		return err
	}
	err = wr.Flush()
	if err != nil {
		return err
	}

	f.Urls = append(f.Urls, urls...)

	return nil
}

// GetAllUserURLs gets all user's short url.
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

// DeleteUserURLs sets the deletion flag to the user URLs sent in the request.
func (f *FileURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
	f.Lock()
	defer f.Unlock()

	for _, delURL := range delURLs {
		for k, curURL := range f.Urls {
			if (delURL == curURL.ShortURL) && (userID == curURL.UserID) {
				f.Urls[k].DeletedFlag = true
				break
			}
		}
	}
	return nil
}

// GetStats gets statistics - amount URLs and users.
func (f *FileURLs) GetStats(ctx context.Context) (stats ServiceStats, err error) {
	f.Lock()
	defer f.Unlock()

	stats.URLs = len(f.Urls)
	stats.Users = 0

	tmp := make([]int, len(f.Urls))
	for _, v := range f.Urls {
		if !slices.Contains(tmp, v.UserID) {
			tmp = append(tmp, v.UserID)
			stats.Users++
		}
	}

	return stats, nil
}

// PingStor checking access to storage.
func (f *FileURLs) PingStor(ctx context.Context) error {
	_, err := os.Stat(f.fileName)
	if os.IsNotExist(err) {
		return errors.New("file not exists")
	}
	return nil
}

// Close closes the storage.
func (f *FileURLs) Close() error {
	f.file.Close()

	var allData []byte
	for _, v := range f.Urls {
		url := FileURL{
			UserID:      v.UserID,
			ShortURL:    v.ShortURL,
			OriginalURL: v.OriginalURL,
			DeletedFlag: v.DeletedFlag,
		}
		data, err := json.Marshal(url)
		if err != nil {
			return err
		}
		allData = append(allData, data...)
		allData = append(allData, '\n')
	}

	newFile, err := os.Create(f.fileName)
	if err != nil {
		return err
	}

	wr := bufio.NewWriter(newFile)
	if _, err = wr.Write(allData); err != nil {
		return err
	}
	err = wr.Flush()
	if err != nil {
		return err
	}

	return newFile.Close()
}
