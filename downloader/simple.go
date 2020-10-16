package downloader

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// Simple ..
type Simple struct {
	Downloader
}

// Download ..
func (d Simple) Download() (reader io.ReadCloser, length int64, err error) {
	var resp *http.Response
	if resp, err = http.Get(d.Task.URL); err != nil {
		return
	}
	reader = resp.Body
	if length, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64); err != nil {
		err = fmt.Errorf("cannot get Content-Length in http header: %v", err)
		return
	}
	return
}
