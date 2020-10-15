package router

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"

	"transfer/database"
)

// Task ..
func Task(app *fiber.App) (err error) {
	var prefix *url.URL
	if prefix, err = url.Parse(viper.GetString("Prefix")); err != nil {
		return
	}

	app.Get("/task", func(ctx *fiber.Ctx) (err error) {
		var tasks []database.Task
		if tasks, err = database.GetTasks(); err != nil {
			return
		}
		for index, content := range tasks {
			tasks[index].Progress = getProgress(content.Name)
		}
		if err = ctx.JSON(tasks); err != nil {
			return
		}
		return
	})

	app.Get("/task/:name", func(ctx *fiber.Ctx) (err error) {
		var task database.Task
		if task, err = database.GetTaskByName(ctx.Params("name")); err != nil {
			return
		}
		task.Progress = getProgress(task.Name)
		if err = ctx.JSON(task); err != nil {
			return
		}
		return
	})

	app.Post("/task", func(ctx *fiber.Ctx) (err error) {
		var task = &database.Task{}
		if err = ctx.BodyParser(task); err != nil {
			ctx.Status(http.StatusBadRequest)
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
				err = fmt.Errorf("cannot get ext name: %s", task.URL)
				ctx.Status(http.StatusBadRequest)
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
				err = fmt.Errorf("url conflict: %s, or you should set force true", task.URL)
				ctx.Status(http.StatusConflict)
				return
			}
		}
		prefix.Path = filepath.Join(task.Path, task.Filename)
		task.DownloadURL = prefix.String()
		task.Status = database.PendingStatus
		if err = task.Insert(); err != nil {
			err = fmt.Errorf("database error: %v", err)
			return
		}
		if err = ctx.JSON(task); err != nil {
			return
		}
		return
	})

	return
}
