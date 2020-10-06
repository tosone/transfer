package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/cheggaaa/pb/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/tosone/logging"
)

type Content struct {
	Type           string          `json:"type"`
	Auth           json.RawMessage `json:"auth"`
	URL            string          `json:"url"`
	Filename       string          `json:"filename"`
	RandomFilename bool            `json:"randomFilename"`
	Path           string          `json:"path"`
	Bucket         string          `json:"bucket"`
	Region         string          `json:"region"`
}

// Qiniu qiniu auth
type Qiniu struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// OSS oss auth
type OSS struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

// Response response
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	var app = fiber.New(fiber.Config{})

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
		if content.RandomFilename {
			var hash = fnv.New32()
			if _, err = hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10))); err != nil {
				return
			}
			var now = time.Now().Format("20060102")
			var id = hex.EncodeToString(hash.Sum(nil))
			var u *url.URL
			if u, err = url.Parse(content.URL); err != nil {
				return
			}
			var ext = filepath.Ext(u.Path)
			if ext == "" {
				response.Message = fmt.Sprintf("cannot get ext name: %s", content.URL)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
				return
			}
			var filename = fmt.Sprintf("%s-%s%s", now, id, ext)
			content.Filename = filename
		}
		switch content.Type {
		case "qiniu":
			var auth = new(Qiniu)
			if err = json.Unmarshal(content.Auth, auth); err != nil {
				response.Code = 1003
				response.Message = fmt.Sprintf("Cannot unmarshal auth: %v", err)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
				return
			}
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
			go func() {
				var err error
				if err = qiniuUpload(auth.AccessKey, auth.SecretKey, content.Region, content.Bucket, length, reader,
					path.Join(content.Path, content.Filename)); err != nil {
					logging.Errorf("Cannot upload the file: %v", err)
					return
				}
				if err = reader.Close(); err != nil {
					logging.Errorf("Close the reader with error: %v", err)
					return
				}
				logging.Infof("Download file success: %v", content.Filename)
			}()
			response.Message = fmt.Sprintf("Downloading %s, content-length: %d", path.Join(content.Path, content.Filename), length)
			if err = c.Status(http.StatusOK).JSON(response); err != nil {
				return
			}
			return
		case "OSS":
			var auth = new(OSS)
			if err = json.Unmarshal(content.Auth, auth); err != nil {
				response.Code = 1003
				response.Message = fmt.Sprintf("Cannot unmarshal auth: %v", err)
				if err = c.Status(http.StatusBadRequest).JSON(response); err != nil {
					return
				}
				return
			}
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
			go func() {
				var err error
				if err = ossUpload(auth.AccessKey, auth.SecretKey, content.Region, content.Bucket, length, reader,
					path.Join(content.Path, content.Filename)); err != nil {
					logging.Errorf("Cannot upload the file: %v", err)
					return
				}
				if err = reader.Close(); err != nil {
					logging.Errorf("Close the reader with error: %v", err)
					return
				}
				logging.Infof("Download file success: %v", content.Filename)
			}()
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

func qiniuUpload(accessKey, secretKey, regionID, bucket string, length int64, reader io.ReadCloser, filename string) (err error) {
	var putPolicy = storage.PutPolicy{
		Scope: bucket,
	}
	var mac = qbox.NewMac(accessKey, secretKey)
	var upToken = putPolicy.UploadToken(mac)
	var region storage.Region
	var exist bool
	if region, exist = storage.GetRegionByID(storage.RegionID(regionID)); !exist {
		err = fmt.Errorf("cannot find the specific region")
		return
	}
	var cfg = storage.Config{
		UseHTTPS:      true,
		UseCdnDomains: true,
		Region:        &region,
	}
	var formUploader = storage.NewFormUploader(&cfg)
	var ret = storage.PutRet{}
	var putExtra = storage.PutExtra{Params: map[string]string{}}

	var bar = pb.Full.Start64(length)
	var barReader = bar.NewProxyReader(reader)
	if err = formUploader.Put(context.Background(), &ret, upToken, filename, barReader, length, &putExtra); err != nil {
		return
	}
	bar.Finish()

	return
}

func ossUpload(accessKey, secretKey, endpoint, bucket string, _ /*length*/ int64, reader io.ReadCloser, filename string) (err error) {
	var client *oss.Client
	if client, err = oss.New(endpoint, accessKey, secretKey); err != nil {
		return
	}
	var bucketObj *oss.Bucket
	if bucketObj, err = client.Bucket(bucket); err != nil {
		return
	}
	if err = bucketObj.PutObject(filename, reader); err != nil {
		return
	}
	return
}
