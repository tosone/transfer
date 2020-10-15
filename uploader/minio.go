package uploader

import (
	"context"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

// Minio ..
type Minio struct {
	Uploader
}

// Upload ..
func (d Minio) Upload() (err error) {
	var endpoint = viper.GetString("minio.endpoint")
	var accessKeyID = viper.GetString("minio.accessKeyID")
	var secretAccessKey = viper.GetString("minio.secretAccessKey")

	var client *minio.Client
	if client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: viper.GetBool("minio.useSSL"),
	}); err != nil {
		return
	}
	var bucket = viper.GetString("minio.bucket")
	var filename = filepath.Join(d.Content.Path, d.Content.Filename)
	if _, err = client.PutObject(context.Background(), bucket, filename, d.Reader,
		d.Length, minio.PutObjectOptions{ContentType: "application/octet-stream"}); err != nil {
		return
	}
	return
}
