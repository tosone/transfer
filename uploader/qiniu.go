package uploader

import (
	"context"
	"fmt"
	"io"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/spf13/viper"
	"github.com/tosone/logging"

	"transfer/database"
)

// Qiniu ..
type Qiniu struct {
	Content database.Content
	Length  int64
	Reader  io.ReadCloser
}

// Upload ..
func (q Qiniu) Upload() (err error) {
	var putPolicy = storage.PutPolicy{
		Scope: viper.GetString("Qiniu.bucket"),
	}

	var mac = qbox.NewMac(viper.GetString("Qiniu.accessKey"), viper.GetString("Qiniu.secretKey"))
	var upToken = putPolicy.UploadToken(mac)
	var region storage.Region
	var exist bool
	if region, exist = storage.GetRegionByID(storage.RegionID(viper.GetString("Qiniu.region"))); !exist {
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
	if err = formUploader.Put(context.Background(), &ret, upToken, q.Content.Filename, q.Reader, q.Length, &putExtra); err != nil {
		logging.Error(err)
	}
	return
}
