package uploader

import (
	"io"
	"transfer/database"
)

type Driver interface {
	Upload() error
}

// Uploader ..
type Uploader struct {
	Content database.Content
	Length  int64
	Reader  io.ReadCloser
}
