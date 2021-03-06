package uploader

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// COS ..
type COS struct {
	Uploader
}

// Upload ..
func (d COS) Upload() (err error) {
	if viper.GetString("COS.region") == "" ||
		viper.GetString("COS.secretId") == "" ||
		viper.GetString("COS.secretKey") == "" {
		err = fmt.Errorf("config is not correct")
		return
	}

	var u *url.URL
	if u, err = url.Parse(viper.GetString("COS.region")); err != nil {
		return
	}

	var baseURL = &cos.BaseURL{BucketURL: u}
	var client = cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  viper.GetString("COS.secretId"),
			SecretKey: viper.GetString("COS.secretKey"),
		},
	})

	var filename = filepath.Join(d.Task.Path, d.Task.Filename)
	if _, err = client.Object.Put(context.Background(), filename, d.Reader, nil); err != nil {
		return
	}
	return
}
