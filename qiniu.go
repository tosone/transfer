package main

import (
	"context"
	"fmt"
	"io"

	"github.com/cheggaaa/pb/v3"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/tosone/logging"
)

// Qiniu ..
type Qiniu struct {
	Content Content
	Task    *Task
	Name    string
	Length  int64
	Reader  io.ReadCloser
}

// Upload ..
func (q Qiniu) Upload() (err error) {
	var putPolicy = storage.PutPolicy{
		Scope: q.Content.Bucket,
	}
	var mac = qbox.NewMac(q.Content.Auth.AccessKey, q.Content.Auth.SecretKey)
	var upToken = putPolicy.UploadToken(mac)
	var region storage.Region
	var exist bool
	if region, exist = storage.GetRegionByID(storage.RegionID(q.Content.Region)); !exist {
		err = fmt.Errorf("cannot find the specific region")
		return
	}
	var cfg = storage.Config{
		UseHTTPS:      true,
		UseCdnDomains: true,
		Region:        &region,
	}
	var formUploader = storage.NewFormUploader(&cfg)
	var ret = storage.PutRet{}
	var putExtra = storage.PutExtra{Params: map[string]string{}}
	var bar = pb.Full.Start64(q.Length)
	DownloadPool.Store(q.Name, DownloadTask{
		URL:         q.Content.URL,
		ProgressBar: bar,
		Filename:    q.Content.Filename,
	})
	var barReader = bar.NewProxyReader(q.Reader)
	go func() {
		var err error
		defer func() {
			if err != nil {
				q.Task.Status = ErrorStatus
				if err := q.Task.UpdateStatus(); err != nil {
					logging.Error(err)
				}
			} else {
				q.Task.Status = DoneStatus
				if err := q.Task.UpdateStatus(); err != nil {
					logging.Error(err)
				}
			}
		}()
		if err = formUploader.Put(context.Background(), &ret, upToken, q.Content.Filename, barReader, q.Length, &putExtra); err != nil {
			logging.Error(err)
		}
		DownloadPool.Delete(q.Name)
		bar.Finish()
		if err = barReader.Close(); err != nil {
			logging.Error(err)
		}
		if err = q.Reader.Close(); err != nil {
			logging.Error(err)
		}
		logging.Infof("Download file success: %v", q.Content.Filename)
	}()
	return
}
