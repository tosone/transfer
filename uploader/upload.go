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
	Task   database.Task
	Length int64
	Reader io.ReadCloser
}
