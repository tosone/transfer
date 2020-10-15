package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"

	"transfer/counter"
	"transfer/database"
	"transfer/notify"
	"transfer/router"
	"transfer/sizewg"
	"transfer/uploader"
)

// MaxTask ..
const MaxTask = 4

// RunTask ..
func RunTask() (err error) {
	var tasks []database.Task
	if tasks, err = database.GetTasksByStatus(database.DoingStatus); err != nil {
		if err != badger.ErrKeyNotFound {
			return
		}
		err = nil
	}
	for _, content := range tasks {
		if err = content.UpdateStatus(database.PendingStatus); err != nil {
			return
		}
	}

	go func() {
		var err error
		for {
			var tasks []database.Task
			if tasks, err = database.GetTasksByStatus(database.PendingStatus); err != nil {
				if err != badger.ErrKeyNotFound {
					logging.Error(err)
				}
				err = nil
				<-time.After(time.Second)
				continue
			}
			for _, task := range tasks {
				DownloadPool.Store(task.Name, task)
				if err = task.UpdateStatus(database.DoingStatus); err != nil {
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
				var task = value.(database.Task)

				if task.Status != database.PendingStatus {
					return true
				}

				taskWaitGroup.Add() // reach to the max task will be blocked

				logging.Infof("task %s is starting: %s", name, task.URL)

				task.Status = database.DoingStatus
				DownloadPool.Store(key, task)

				go func() {
					defer taskWaitGroup.Done()
					defer DownloadPool.Delete(task.Name)

					if err = TaskHandler(task); err != nil {
						logging.Error(err)
						if err = task.UpdateStatus(database.ErrorStatus); err != nil {
							logging.Error(err)
						}
					} else {
						if err = task.UpdateStatus(database.DoneStatus); err != nil {
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

// TaskHandler ..
func TaskHandler(task database.Task) (err error) {
	var reader io.ReadCloser
	var length int64
	if reader, length, err = downloader(task.URL); err != nil {
		return
	}

	var proxyReader *counter.Reader
	if proxyReader, err = counter.Proxy(reader, length); err != nil {
		return
	}
	router.ProgressBarPool.Store(task.Name, proxyReader)
	defer func() {
		router.ProgressBarPool.Delete(task.Name)
		if err := proxyReader.Close(); err != nil {
			logging.Error(err)
		}
		if err := reader.Close(); err != nil {
			logging.Error(err)
		}
	}()

	var driver uploader.Driver
	var upload = uploader.Uploader{
		Task:   task,
		Length: length,
		Reader: proxyReader,
	}
	switch task.Type {
	case "qiniu":
		driver = uploader.Qiniu{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "OSS":
		driver = uploader.OSS{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "COS":
		driver = uploader.COS{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "minio":
		driver = uploader.Minio{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "s3":
		driver = uploader.S3{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	case "local":
		driver = uploader.Local{Uploader: upload}
		if err = driver.Upload(); err != nil {
			return
		}
	default:
		err = fmt.Errorf("not support this type: %s", task.Type)
		return
	}
	logging.Infof("download file success: %v", task.Filename)

	if com.IsSliceContainsStr(viper.GetStringSlice("Notify"), "Email") {
		if err = notify.Mail(task); err != nil {
			return
		}
	}
	if com.IsSliceContainsStr(viper.GetStringSlice("Notify"), "SMS") {
		if err = notify.SMS(task); err != nil {
			return
		}
	}

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
