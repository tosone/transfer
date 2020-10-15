package uploader

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/qiniu/api.v7/v7/auth/qbox"
	"github.com/qiniu/api.v7/v7/storage"
	"github.com/spf13/viper"
)

// Qiniu ..
type Qiniu struct {
	Uploader
}

// Upload ..
func (d Qiniu) Upload() (err error) {
	var putPolicy = storage.PutPolicy{
		Scope: viper.GetString("qiniu.bucket"),
	}

	var mac = qbox.NewMac(viper.GetString("qiniu.accessKey"), viper.GetString("qiniu.secretKey"))
	var upToken = putPolicy.UploadToken(mac)
	var region storage.Region
	var exist bool
	if region, exist = storage.GetRegionByID(storage.RegionID(viper.GetString("qiniu.region"))); !exist {
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
	var filename = filepath.Join(d.Content.Path, d.Content.Filename)
	if err = formUploader.Put(context.Background(), &ret, upToken, filename,
		d.Reader, d.Length, &putExtra); err != nil {
		return
	}
	return
}
