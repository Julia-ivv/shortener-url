// Package deleter uses channels to async delete a list of user's URLs.
package deleter

import (
	"database/sql"
	"sync"

	"github.com/Julia-ivv/shortener-url.git/pkg/logger"
)

// NumWorkers - number of channels receiving data for deletion.
const NumWorkers = 5

// DelData contains the URL for deletion and its user ID.
type DelData struct {
	ShortURL string
	UserID   int
}

// Generator sends URLs to the channel for deletion.
func Generator(doneCh chan struct{}, input []string, userID int) chan DelData {
	inputCh := make(chan DelData)

	go func() {
		defer close(inputCh)

		for _, url := range input {
			select {
			case <-doneCh:
				return
			case inputCh <- DelData{ShortURL: url, UserID: userID}:
			}
		}
	}()

	return inputCh
}

// DelResult contains the result of removing the URL.
type DelResult struct {
	Err  error
	Rows int64
}

// del receives the URL for deletion from the channel inputCh,
// marks the URL as deleted in the storage,
// places the result of the work in the channel delRes.
func del(doneCh chan struct{}, inputCh chan DelData, stmt *sql.Stmt) chan DelResult {
	delRes := make(chan DelResult)

	go func() {
		defer close(delRes)
		for data := range inputCh {
			result, errEx := stmt.Exec(data.UserID, data.ShortURL)
			rows, err := result.RowsAffected()
			if err != nil {
				logger.ZapSugar.Infow("returns the number of rows", err)
			}
			select {
			case <-doneCh:
				return
			case delRes <- DelResult{Rows: rows, Err: errEx}:
			}
		}
	}()

	return delRes
}

// FanOut fills the channel with the deletion results.
func FanOut(doneCh chan struct{}, inputCh chan DelData, stmt *sql.Stmt) []chan DelResult {
	channels := make([]chan DelResult, NumWorkers)
	for i := 0; i < NumWorkers; i++ {
		delResutlCh := del(doneCh, inputCh, stmt)
		channels[i] = delResutlCh
	}
	return channels
}

// FanIn collects URL removal results from several channels into one.
func FanIn(stmt *sql.Stmt, doneCh chan struct{}, chans ...chan DelResult) chan DelResult {
	finalCh := make(chan DelResult)
	var wGroup sync.WaitGroup

	for _, ch := range chans {
		iterChan := ch
		wGroup.Add(1)

		go func() {
			defer wGroup.Done()
			for data := range iterChan {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wGroup.Wait()
		close(finalCh)
	}()

	return finalCh
}
