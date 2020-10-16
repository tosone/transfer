package downloader

import (
	"fmt"
	"io"
	"transfer/database"
)

// Driver ..
type Driver interface {
	Download() (io.ReadCloser, int64, error)
}

// Downloader ..
type Downloader struct {
	Task database.Task
}

// Download ..
func Download(downloadType string, task database.Task) (reader io.ReadCloser, length int64, err error) {
	var driver Driver
	switch downloadType {
	case "simple":
		driver = Simple{Downloader: Downloader{Task: task}}
		if reader, length, err = driver.Download(); err != nil {
			return
		}
	default:
		err = fmt.Errorf("not support this download type: %s", downloadType)
		return
	}
	return
}
