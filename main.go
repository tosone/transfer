package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
	"transfer/database"

	"github.com/cheggaaa/pb/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/tosone/logging"
	"gorm.io/gorm"
)

func main() {
	var err error

	if err = database.Database(); err != nil {
		logging.Fatal(err)
	}

	RunTask()

	if err = Config(); err != nil {
		logging.Fatal(err)
	}

	var app = fiber.New()

	app.Use(compress.New())
	app.Use(requestid.New())

	app.Get("/status", func(c *fiber.Ctx) (err error) {
		var result = make(map[string]database.Content)
		DownloadPool.Range(func(key, value interface{}) bool {
			var content = value.(database.Content)
			var name = key.(string)
			ProgressBarPool.Range(func(key, value interface{}) bool {
				if key.(string) == name {
					var bar = value.(*pb.ProgressBar)
					var progress = fmt.Sprintf("%.2f", float64(bar.Current()*100.0)/float64(bar.Total()))
					content.Progress = progress
					return false
				}
				return true
			})
			return true
		})
		if err = c.Status(http.StatusOK).JSON(result); err != nil {
			return
		}
		return
	})

	app.Get("/status/:name", func(ctx *fiber.Ctx) (err error) {
		var content database.Content
		DownloadPool.Range(func(key, value interface{}) bool {
			content = value.(database.Content)
			var name = key.(string)
			if name == ctx.Params("name") {
				ProgressBarPool.Range(func(key, value interface{}) bool {
					if key.(string) == name {
						var bar = value.(*pb.ProgressBar)
						var progress = fmt.Sprintf("%.2f", float64(bar.Current()*100.0)/float64(bar.Total()))
						content.Progress = progress
						return false
					}
					return true
				})
				return false
			}
			return true
		})
		if err = ctx.Status(http.StatusOK).JSON(content); err != nil {
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
			var task database.Content
			if err = database.DBEngine.Where(&database.Content{URL: content.URL}).First(&task).Error; err == gorm.ErrRecordNotFound {
				err = nil
			} else if err != nil {
				err = fmt.Errorf("database error: %s", content.URL)
				ctx.Status(http.StatusBadRequest)
				return
			} else {
				err = fmt.Errorf("url conflict: %s, or you should set force true", content.URL)
				ctx.Status(http.StatusConflict)
				return
			}
		}
		var contentBytes []byte
		if contentBytes, err = json.Marshal(content); err != nil {
			return
		}
		content.Content = contentBytes
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

	if err = app.Listen(":3000"); err != nil {
		logging.Fatal(err)
	}
}
