package uploader

import (
	"context"
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

	var filename = filepath.Join(d.Content.Path, d.Content.Filename)
	if _, err = client.Object.Put(context.Background(), filename, d.Reader, nil); err != nil {
		return
	}
	return
}
