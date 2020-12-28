package main

import (
	"io"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"

	"transfer/counter"
	"transfer/database"
	"transfer/downloader"
	"transfer/notify"
	"transfer/router"
	"transfer/sizewg"
	"transfer/uploader"
)

var stopWaitGroup = &sync.WaitGroup{}

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

	stopWaitGroup.Add(1)
	var taskWaitGroup = sizewg.New(viper.GetInt("Config.ParallelTask"))
	go func() {
		defer stopWaitGroup.Done()
		var err error
		for !appStopped {
			var tasks []database.Task
			if tasks, err = database.GetTasksByStatus(database.PendingStatus); err != nil {
				if err != badger.ErrKeyNotFound {
					logging.Error(err)
				}
				if err == badger.ErrDBClosed {
					break
				}
				err = nil
				<-time.After(time.Second)
				continue
			}
			for _, task := range tasks {
				taskWaitGroup.Add() // reach to the max task will be blocked
				stopWaitGroup.Add(1)
				logging.Infof("task %s is starting: %s", task.Name, task.URL)
				if err = task.UpdateStatus(database.DoingStatus); err != nil {
					logging.Error(err)
				}
				go func(task database.Task) {
					defer taskWaitGroup.Done()
					defer stopWaitGroup.Done()

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
				}(task)
			}
			<-time.After(time.Second)
		}
	}()

	return
}

// TaskHandler ..
func TaskHandler(task database.Task) (err error) {
	var reader io.ReadCloser
	var length int64
	if reader, length, err = downloader.Download(task.DownloadType, task); err != nil {
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

	var upload = uploader.Uploader{
		Task:   task,
		Length: length,
		Reader: proxyReader,
	}
	if err = uploader.Upload(task.UploadType, upload); err != nil {
		return
	}
	logging.Infof("download file success: %v", task.Filename)

	if com.IsSliceContainsStr(viper.GetStringSlice("Notify"), "Email") {
		if err = notify.Mail(task, notify.DownloadSuccess); err != nil {
			return
		}
	}

	return
}
