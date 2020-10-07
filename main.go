package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/tosone/logging"
)

type Content struct {
	Type           string `json:"type"`
	Auth           Auth   `json:"auth"`
	URL            string `json:"url"`
	Filename       string `json:"filename"`
	RandomFilename bool   `json:"randomFilename"`
	Path           string `json:"path"`
	Bucket         string `json:"bucket"`
	Region         string `json:"region"`
	Endpoint       string `json:"endpoint"`
	Force          bool   `json:"force"`
}

// Auth auth
type Auth struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// Response response
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	var app = fiber.New(fiber.Config{Prefork: true})

	app.Get("/status", func(c *fiber.Ctx) (err error) {
		var result = make(map[string]DownloadTask)
		DownloadPool.Range(func(key, value interface{}) bool {
			var downloadTask = value.(DownloadTask)
			var bar = downloadTask.ProgressBar
			var progress = fmt.Sprintf("%.2f", float64(bar.Current()*100.0)/float64(bar.Total()))
			downloadTask.Progress = fmt.Sprintf("%s%%", progress)
			result[key.(string)] = downloadTask
			return true
		})
		if err = c.Status(http.StatusOK).JSON(result); err != nil {
			return
		}
		return
	})

	app.Post("/download", func(c *fiber.Ctx) (err error) {
		var content Content
		var response = Response{Code: 200, Message: "Success"}
		if err = c.BodyParser(&content); err != nil {
			response.Code = 1001
			response.Message = err.Error()
			if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
				return
			}
			return
		}
		var name string
		if name, err = GenName(); err != nil {
			return
		}

		if content.RandomFilename {
			var now = time.Now().Format("20060102")
			var u *url.URL
			if u, err = url.Parse(content.URL); err != nil {
				return
			}
			var ext = filepath.Ext(u.Path)
			if ext == "" {
				response.Code = 1003
				response.Message = fmt.Sprintf("cannot get ext name: %s", content.URL)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
				return
			}
			var filename = fmt.Sprintf("%s-%s%s", now, name, ext)
			content.Filename = filename
		}

		if !content.Force {
			var task Task
			if err = DBEngine.Where(&Task{URL: content.URL}).First(&task).Error; err == gorm.ErrRecordNotFound {
				err = nil
			} else if err != nil {
				response.Code = 1003
				response.Message = fmt.Sprintf("database error: %s", content.URL)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
			} else {
				response.Code = 1003
				response.Message = fmt.Sprintf("url conflict: %s, or you should set force true", content.URL)
				if err = c.Status(http.StatusConflict).JSON(response); err != nil {
					return
				}
				return
			}
		}

		var task = &Task{
			Name:     name,
			URL:      content.URL,
			Filename: content.Filename,
			Status:   DoingStatus,
		}
		if err = task.Insert(); err != nil {
			response.Code = 1003
			response.Message = fmt.Sprintf("Database error: %v", err)
			if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
				return
			}
			return
		}
		defer func() {
			if err != nil || response.Code != 200 {
				task.Status = ErrorStatus
				if err := task.UpdateStatus(); err != nil {
					logging.Error(err)
				}
			}
		}()

		var reader io.ReadCloser
		var length int64
		if reader, length, err = downloader(content.URL); err != nil {
			response.Code = 1003
			response.Message = fmt.Sprintf("Cannot download the file: %v", err)
			if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
				return
			}
			return
		}

		switch content.Type {
		case "qiniu":
			var qiniu = Qiniu{
				Content: content,
				Task:    task,
				Name:    name,
				Length:  length,
				Reader:  reader,
			}
			if err = qiniu.Upload(); err != nil {
				response.Code = 1003
				response.Message = fmt.Sprintf("Upload file with error: %v", err)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
				return
			}
			response.Message = fmt.Sprintf("Downloading %s, content-length: %d", path.Join(content.Path, content.Filename), length)
			if err = c.Status(http.StatusOK).JSON(response); err != nil {
				return
			}
			return
		case "OSS":
			var oss = OSS{
				Content: content,
				Task:    task,
				Name:    name,
				Length:  length,
				Reader:  reader,
			}
			if err = oss.Upload(); err != nil {
				response.Code = 1003
				response.Message = fmt.Sprintf("Upload file with error: %v", err)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
				return
			}
			response.Message = fmt.Sprintf("Downloading %s, content-length: %d", path.Join(content.Path, content.Filename), length)
			if err = c.Status(http.StatusOK).JSON(response); err != nil {
				return
			}
			return
		default:
			response.Code = 1002
			response.Message = "Not support this kind of storage"
			if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
				return
			}
			return
		}
	})

	var err error

	if err = Database(); err != nil {
		logging.Fatal(err)
	}

	if err = app.Listen(":3000"); err != nil {
		logging.Fatal(err)
	}
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
