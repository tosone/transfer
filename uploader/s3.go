package uploader

import (
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
)

// S3 ..
type S3 struct {
	Uploader
}

// Upload ..
func (d S3) Upload() (err error) {
	var config = &aws.Config{
		Region: aws.String("s3.region"),
		Credentials: credentials.NewStaticCredentials(
			viper.GetString("s3.accessKeyID"),
			viper.GetString("s3.secretAccessKey"),
			"",
		),
	}

	var sess *session.Session
	if sess, err = session.NewSession(config); err != nil {
		return
	}

	var object = &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(d.Reader),
		Bucket: aws.String(viper.GetString("s3.bucket")),
		Key:    aws.String(filepath.Join(d.Task.Path, d.Task.Filename)),
	}

	var svc = s3.New(sess)
	if _, err = svc.PutObject(object); err != nil {
		return
	}

	return
}
