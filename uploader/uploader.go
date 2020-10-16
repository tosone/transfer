package uploader

import (
	"fmt"
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

func Upload(uploadType string, upload Uploader) (err error) {
	var driver Driver
	switch uploadType {
	case "qiniu":
		driver = Qiniu{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "OSS":
		driver = OSS{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "COS":
		driver = COS{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "minio":
		driver = Minio{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "s3":
		driver = S3{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "local":
		driver = Local{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	default:
		err = fmt.Errorf("not support this type: %s", uploadType)
		return
	}
	return
}
