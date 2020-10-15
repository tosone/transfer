package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/dgraph-io/badger/v2"
	"github.com/tosone/logging"

	"transfer/database"
	"transfer/sizewg"
	"transfer/uploader"
)

const MaxTask = 4

func RunTask() (err error) {
	var contents []database.Content
	if contents, err = database.GetContentsByStatus(database.DoingStatus); err != nil {
		if err != badger.ErrKeyNotFound {
			return
		}
		err = nil
	}
	for _, content := range contents {
		if err = content.UpdateStatus(database.PendingStatus); err != nil {
			return
		}
	}

	go func() {
		var err error
		for {
			var contents []database.Content
			if contents, err = database.GetContentsByStatus(database.PendingStatus); err != nil {
				if err != badger.ErrKeyNotFound {
					logging.Error(err)
				}
				err = nil
				<-time.After(time.Second)
				continue
			}
			for _, content := range contents {
				DownloadPool.Store(content.Name, content)
				if err = content.UpdateStatus(database.DoingStatus); err != nil {
					logging.Error(err)
				}
			}
			<-time.After(time.Second)
		}
	}()

	var taskWaitGroup = sizewg.New(MaxTask)

	go func() {
		var err error
		for {
			DownloadPool.Range(func(key, value interface{}) bool {
				var name = key.(string)
				var content = value.(database.Content)

				if content.Status != database.PendingStatus {
					return true
				}

				taskWaitGroup.Add() // reach to the max task will be blocked

				logging.Infof("task %s is starting: %s", name, content.URL)

				content.Status = database.DoingStatus
				DownloadPool.Store(key, content)

				go func() {
					defer taskWaitGroup.Done()
					defer DownloadPool.Delete(content.Name)

					if err = TaskHandler(content); err != nil {
						logging.Error(err)
						if err = content.UpdateStatus(database.ErrorStatus); err != nil {
							logging.Error(err)
						}
					} else {
						if err = content.UpdateStatus(database.DoneStatus); err != nil {
							logging.Error(err)
						}
					}
				}()

				return true
			})
			<-time.After(time.Second)
		}
	}()

	return
}

func TaskHandler(content database.Content) (err error) {
	var reader io.ReadCloser
	var length int64
	if reader, length, err = downloader(content.URL); err != nil {
		return
	}

	var bar = pb.Full.Start64(length)
	ProgressBarPool.Store(content.Name, bar)
	var barReader = bar.NewProxyReader(reader)
	defer func() {
		ProgressBarPool.Delete(content.Name)
		bar.Finish()
		if err := barReader.Close(); err != nil {
			logging.Error(err)
		}
		if err := reader.Close(); err != nil {
			logging.Error(err)
		}
	}()

	var driver uploader.Driver
	var upload = uploader.Uploader{
		Content: content,
		Length:  length,
		Reader:  barReader,
	}
	switch content.Type {
	case "qiniu":
		driver = uploader.Qiniu{
			Uploader: upload,
		}
		if err = driver.Upload(); err != nil {
			return
		}
	case "OSS":
		driver = uploader.OSS{
			Uploader: upload,
		}
		if err = driver.Upload(); err != nil {
			return
		}
		return
	case "COS":
		driver = uploader.COS{
			Uploader: upload,
		}
		if err = driver.Upload(); err != nil {
			return
		}
		return
	default:
		err = fmt.Errorf("not support this type: %s", content.Type)
		return
	}
	logging.Infof("download file success: %v", content.Filename)
	return
}

func downloader(url string) (reader io.ReadCloser, length int64, err error) {
	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return
	}
	reader = resp.Body
	if length, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64); err != nil {
		err = fmt.Errorf("cannot get Content-Length in http header: %v", err)
		return
	}
	return
}
