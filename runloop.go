package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"transfer/sizewg"

	"github.com/cheggaaa/pb/v3"
	"github.com/tosone/logging"
)

const MaxTask = 4

func RunTask() {
	var err error

	go func() {
		for {
			var tasks []Content
			if err = DBEngine.Where(Content{Status: PendingStatus}).Find(&tasks).Error; err != nil {
				logging.Error(err)
			}
			for _, task := range tasks {
				var content Content
				if err = json.Unmarshal(task.Content, &content); err != nil {
					logging.Error(err)
				}
				DownloadPool.Store(task.Name, content)
			}
			<-time.After(time.Second)
		}
	}()

	var taskWaitGroup = sizewg.New(MaxTask)

	go func() {
		for {
			DownloadPool.Range(func(key, value interface{}) bool {
				taskWaitGroup.Add() // reach to the max task will be blocked
				go func() {
					defer taskWaitGroup.Done()
					var content = value.(Content)
					var name = key.(string)
					if err = TaskHandler(content); err != nil {
						if err = DBEngine.Model(Content{}).Where(Content{Name: name}).
							Updates(Content{Status: ErrorStatus, Message: err.Error()}).Error; err != nil {
							logging.Error(err)
						}
					} else {
						if err = DBEngine.Model(Content{}).Where(Content{Name: name}).
							Updates(Content{Status: DoneStatus}).Error; err != nil {
							logging.Error(err)
						}
					}
				}()
				return true
			})
			<-time.After(time.Second)
		}
	}()
}

func TaskHandler(content Content) (err error) {
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

	var driver Driver
	switch content.Type {
	case "qiniu":
		driver = Qiniu{
			Content: content,
			Length:  length,
			Reader:  barReader,
		}
		if err = driver.Upload(); err != nil {
			return
		}
	case "OSS":
		driver = OSS{
			Content: content,
			Length:  length,
			Reader:  barReader,
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
