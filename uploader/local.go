package uploader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/unknwon/com"
)

// Local ..
type Local struct {
	Uploader
}

// Upload ..
func (d Local) Upload() (err error) {
	var dir = viper.GetString("local.dir")
	var filename = filepath.Join(dir, d.Task.Path, d.Task.Filename)
	if com.IsFile(filename) {
		err = fmt.Errorf("file already exist: %s", filename)
		return
	}
	if err = os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return
	}
	var file *os.File
	if file, err = os.Create(filename); err != nil {
		return
	}
	if _, err = io.Copy(file, d.Reader); err != nil {
		return
	}
	return
}
