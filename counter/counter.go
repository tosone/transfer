package counter

import (
	"io"
	"sync"

	"github.com/tosone/logging"
)

type Counter struct {
	length  int64
	current int64

	Percent float64
}

// Reader ..
type Reader struct {
	data         []byte
	readerLocker *sync.Mutex
	eof          bool

	length  int64
	current int64

	Percent float64
}

// Read ..
func (r *Reader) Read(data []byte) (n int, err error) {
	r.readerLocker.Lock()
	defer r.readerLocker.Unlock()
	n = copy(data, r.data[:])
	r.data = r.data[n:]
	if n == 0 && r.eof {
		err = io.EOF
	}
	return
}

// Close ..
func (r *Reader) Close() (err error) {
	return
}

func Proxy(reader io.ReadCloser, length int64) (result *Reader, err error) {
	result = &Reader{
		length:       length,
		readerLocker: new(sync.Mutex),
	}
	go func() {
		for {
			var data = make([]byte, 512)
			var n int
			n, err = reader.Read(data)
			result.readerLocker.Lock()
			result.data = append(result.data, data[:n]...)
			result.current += int64(n)
			result.Percent = float64(result.current) * 100 / float64(result.length)
			result.readerLocker.Unlock()
			if err == io.EOF {
				result.eof = true
				break
			}
			if err != nil {
				logging.Error(err)
				break
			}
		}
	}()
	return
}
