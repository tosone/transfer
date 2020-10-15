package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/dgraph-io/badger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
	"github.com/unknwon/com"

	"transfer/database"
)

// DownloadPool ..
var DownloadPool = sync.Map{}

// ProgressBarPool ..
var ProgressBarPool = sync.Map{}

// ConfigFile default config path
const ConfigFile = "/etc/transfer/config.yaml"

func Config() (err error) {
	var configFile string
	flag.StringVar(&configFile, "config", ConfigFile, "config file")
	flag.Parse()

	if !com.IsFile(configFile) {
		logging.Fatalf("cannot find config file: %s", configFile)
	}
	viper.SetConfigType("yaml")
	viper.SetConfigName(filepath.Base(configFile))
	viper.AddConfigPath(filepath.Dir(configFile))

	if err = viper.ReadInConfig(); err != nil {
		return
	}
	return
}

func main() {
	var err error

	if err = Config(); err != nil {
		logging.Fatal(err)
	}

	if err = database.Initialize(); err != nil {
		logging.Fatal(err)
	}
	defer func() {
		if err = database.Teardown(); err != nil {
			logging.Error(err)
		}
	}()

	if err = RunTask(); err != nil {
		logging.Fatal(err)
	}

	var prefix *url.URL
	if prefix, err = url.Parse(viper.GetString("Prefix")); err != nil {
		logging.Fatal(err)
	}

	var app = fiber.New()

	app.Use(compress.New())
	app.Use(requestid.New())

	app.Get("/task", func(ctx *fiber.Ctx) (err error) {
		var tasks []database.Task
		if tasks, err = database.GetContents(); err != nil {
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
		if task, err = database.GetContentByName(ctx.Params("name")); err != nil {
			return
		}
		task.Progress = getProgress(task.Name)
		if err = ctx.JSON(task); err != nil {
			return
		}
		return
	})

	app.Post("/download", func(ctx *fiber.Ctx) (err error) {
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
			if _, err = database.GetContentByURL(task.URL); err != nil {
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

	go func() {
		if err = app.Listen(":3000"); err != nil {
			logging.Fatal(err)
		}
	}()

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, os.Interrupt)

	<-signalChanel

	logging.Info("transfer has been stopped")
}

func getProgress(name string) (progress string) {
	ProgressBarPool.Range(func(key, value interface{}) bool {
		if key.(string) == name {
			var bar = value.(*pb.ProgressBar)
			progress = fmt.Sprintf("%.2f", float64(bar.Current()*100.0)/float64(bar.Total()))
			return false
		}
		return true
	})
	return
}
