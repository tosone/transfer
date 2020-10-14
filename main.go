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
		var contents []database.Content
		if contents, err = database.GetContents(); err != nil {
			return
		}
		for index, content := range contents {
			contents[index].Progress = getProgress(content.Name)
		}
		if err = ctx.JSON(contents); err != nil {
			return
		}
		return
	})

	app.Get("/task/:name", func(ctx *fiber.Ctx) (err error) {
		var content database.Content
		if content, err = database.GetContentByName(ctx.Params("name")); err != nil {
			return
		}
		content.Progress = getProgress(content.Name)
		if err = ctx.JSON(content); err != nil {
			return
		}
		return
	})

	app.Post("/download", func(ctx *fiber.Ctx) (err error) {
		var content = &database.Content{}
		if err = ctx.BodyParser(content); err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
		var name string
		if name, err = database.GenName(); err != nil {
			return
		}
		content.Name = name
		if content.RandomFilename {
			var now = time.Now().Format("20060102")
			var u *url.URL
			if u, err = url.Parse(content.URL); err != nil {
				return
			}
			var ext = filepath.Ext(u.Path)
			if ext == "" {
				err = fmt.Errorf("cannot get ext name: %s", content.URL)
				ctx.Status(http.StatusBadRequest)
				return
			}
			var filename = fmt.Sprintf("%s-%s%s", now, name, ext)
			content.Filename = filename
		}
		if !content.Force {
			if _, err = database.GetContentByURL(content.URL); err != nil {
				if err != badger.ErrKeyNotFound {
					return
				}
				err = nil
			} else {
				err = fmt.Errorf("url conflict: %s, or you should set force true", content.URL)
				ctx.Status(http.StatusConflict)
				return
			}
		}
		prefix.Path = filepath.Join(content.Path, content.Filename)
		content.DownloadURL = prefix.String()
		content.Status = database.PendingStatus
		if err = content.Insert(); err != nil {
			err = fmt.Errorf("database error: %v", err)
			return
		}
		if err = ctx.JSON(content); err != nil {
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
