package router

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"transfer/database"
)

// Task ..
func Task(app *gin.Engine) (err error) {
	var prefix *url.URL
	if prefix, err = url.Parse(viper.GetString("Prefix")); err != nil {
		return
	}

	app.GET("/task", func(ctx *gin.Context) {
		var tasks []database.Task
		var status = database.Status(ctx.Query("status"))
		if status != "" {
			if tasks, err = database.GetTasksByStatus(status); err != nil {
				_ = ctx.Error(errDatabase.Build(err))
				return
			}
		} else {
			if tasks, err = database.GetTasks(); err != nil {
				_ = ctx.Error(errDatabase.Build(err))
				return
			}
		}
		for index, content := range tasks {
			if content.Status == database.DoneStatus {
				tasks[index].Progress = 100
			} else {
				if tasks[index].Progress, err = getProgress(content.Name); err != nil {
					_ = ctx.Error(errServerInternal.Build(err))
					return
				}
			}
		}
		ctx.JSON(http.StatusOK, tasks)
		return
	})

	app.GET("/task/:name", func(ctx *gin.Context) {
		var task database.Task
		if task, err = database.GetTaskByName(ctx.Param("name")); err != nil {
			_ = ctx.Error(errDatabase.Build(err))
			return
		}
		if task.Progress, err = getProgress(task.Name); err != nil {
			_ = ctx.Error(errServerInternal.Build(err))
			return
		}
		ctx.JSON(http.StatusOK, task)
		return
	})

	app.POST("/task", func(ctx *gin.Context) {
		var task = &database.Task{}
		if err = ctx.Bind(task); err != nil {
			_ = ctx.Error(errBadRequest.Build(err))
			return
		}
		var name string
		if name, err = database.GenName(); err != nil {
			return
		}
		task.Name = name
		if task.RandomFilename {
			var now = time.Now().Format("20060102")
			var u *url.URL
			if u, err = url.Parse(task.URL); err != nil {
				return
			}
			var ext = filepath.Ext(u.Path)
			if ext == "" {
				_ = ctx.Error(errBadRequest.Build(fmt.Sprintf("cannot get ext name %s", task.URL)))
				return
			}
			var filename = fmt.Sprintf("%s-%s%s", now, name, ext)
			task.Filename = filename
		}
		if !task.Force {
			if _, err = database.GetTaskByURL(task.URL); err != nil {
				if err != badger.ErrKeyNotFound {
					return
				}
				err = nil
			} else {
				_ = ctx.Error(errBadRequest.Build(errURLConflict.Build(task.URL)))
				return
			}
		}
		prefix.Path = filepath.Join(task.Path, task.Filename)
		task.DownloadURL = prefix.String()
		task.Status = database.PendingStatus
		if err = task.Insert(); err != nil {
			_ = ctx.Error(errDatabase.Build(err))
			return
		}
		ctx.JSON(http.StatusOK, task)
		return
	})

	return
}
