package main

import (
	"io"

	"github.com/tosone/logging"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cheggaaa/pb/v3"
)

// OSS ..
type OSS struct {
	Content Content
	Task    *Task
	Name    string
	Length  int64
	Reader  io.ReadCloser
}

// Upload ..
func (o *OSS) Upload() (err error) {
	var client *oss.Client
	if client, err = oss.New(o.Content.Endpoint, o.Content.Auth.AccessKey, o.Content.Auth.SecretKey); err != nil {
		return
	}
	var bucketObj *oss.Bucket
	if bucketObj, err = client.Bucket(o.Content.Bucket); err != nil {
		return
	}
	var bar = pb.Full.Start64(o.Length)
	DownloadPool.Store(o.Name, DownloadTask{
		URL:         o.Content.URL,
		ProgressBar: bar,
		Filename:    o.Content.Filename,
	})
	var barReader = bar.NewProxyReader(o.Reader)
	go func() {
		var err error
		defer func() {
			if err != nil {
				o.Task.Status = ErrorStatus
				if err := o.Task.UpdateStatus(); err != nil {
					logging.Error(err)
				}
			} else {
				o.Task.Status = DoneStatus
				if err := o.Task.UpdateStatus(); err != nil {
					logging.Error(err)
				}
			}
		}()
		if err = bucketObj.PutObject(o.Content.Filename, barReader); err != nil {
			logging.Error(err)
		}
		DownloadPool.Delete(o.Name)
		bar.Finish()
		if err = barReader.Close(); err != nil {
			logging.Error(err)
		}
		if err = o.Reader.Close(); err != nil {
			logging.Error(err)
		}
		logging.Infof("Download file success: %v", o.Content.Filename)
	}()
	return
}
