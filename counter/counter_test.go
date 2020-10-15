package counter

import (
	"io"
	"os"
	"testing"
)

func TestProxy(t *testing.T) {
	var err error

	var file *os.File
	if file, err = os.Open("counter.go"); err != nil {
		t.Error(err)
	}
	var reader *Reader
	if reader, err = Proxy(file, 0); err != nil {
		t.Error(err)
	}

	var file1 *os.File
	if file1, err = os.Create("test.data"); err != nil {
		t.Error(err)
	}
	if _, err = io.Copy(file1, reader); err != nil {
		t.Error(err)
	}

	if err = file1.Close(); err != nil {
		t.Error(err)
	}
	if err = file.Close(); err != nil {
		t.Error(err)
	}
}
