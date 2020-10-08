package main

import (
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
)

// OSS ..
type OSS struct {
	Content Content
	Length  int64
	Reader  io.ReadCloser
}

// Upload ..
func (o OSS) Upload() (err error) {
	var client *oss.Client
	if client, err = oss.New(o.Content.Endpoint, viper.GetString("OSS.accessKey"), viper.GetString("OSS.secretKey")); err != nil {
		return
	}
	var bucketObj *oss.Bucket
	if bucketObj, err = client.Bucket(o.Content.Bucket); err != nil {
		return
	}
	if err = bucketObj.PutObject(o.Content.Filename, o.Reader); err != nil {
		logging.Error(err)
	}
	return
}
