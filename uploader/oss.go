package uploader

import (
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
	"github.com/tosone/logging"
)

// OSS ..
type OSS struct {
	Uploader
}

// Upload ..
func (d OSS) Upload() (err error) {
	var client *oss.Client
	if client, err = oss.New(viper.GetString("OSS.endpoint"),
		viper.GetString("OSS.accessKey"), viper.GetString("OSS.secretKey")); err != nil {
		return
	}
	var bucketObj *oss.Bucket
	if bucketObj, err = client.Bucket(viper.GetString("OSS.bucket")); err != nil {
		return
	}
	var filename = filepath.Join(d.Content.Path, d.Content.Filename)
	if err = bucketObj.PutObject(filename, d.Reader); err != nil {
		logging.Error(err)
	}
	return
}
