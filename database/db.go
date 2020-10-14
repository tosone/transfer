package database

import (
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
)

// Status ..
type Status string

const (
	// DoneStatus ..
	DoneStatus Status = "done"
	// ErrorStatus ..
	ErrorStatus Status = "error"
	// PendingStatus ..
	PendingStatus Status = "pending"
	// DoingStatus ..
	DoingStatus Status = "doing"
)

type Content struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	URL            string `json:"url"`
	DownloadURL    string `json:"downloadUrl"`
	Filename       string `json:"filename"`
	RandomFilename bool   `json:"randomFilename"`
	Path           string `json:"path"`
	Force          bool   `json:"force"`
	Progress       string `json:"progress"`
	Status         Status `json:"status"`
	Message        string `json:"message"`
}

var dbEngine *badger.DB

// Initialize ..
func Initialize() (err error) {
	if dbEngine, err = badger.Open(badger.DefaultOptions(viper.GetString("DatabaseDir"))); err != nil {
		return
	}
	return
}

// Teardown ..
func Teardown() (err error) {
	if err = dbEngine.Close(); err != nil {
		return
	}
	return
}

// Insert ..
func (t *Content) Insert() (err error) {
	var txn = dbEngine.NewTransaction(true)

	var data []byte
	if data, err = json.Marshal(*t); err != nil {
		return
	}

	var entry = badger.NewEntry([]byte(t.Name), data)
	if err = txn.SetEntry(entry); err != nil {
		return
	}
	if err = txn.Commit(); err != nil {
		return
	}
	return
}

// UpdateStatus ..
func (t *Content) UpdateStatus(status Status) (err error) {
	var txn = dbEngine.NewTransaction(true)
	defer func() {
		if err = txn.Commit(); err != nil {
			return
		}
	}()

	t.Status = status

	var data []byte
	if data, err = json.Marshal(*t); err != nil {
		return
	}

	var entry = badger.NewEntry([]byte(t.Name), data)
	if err = txn.SetEntry(entry); err != nil {
		return
	}

	return
}

// GetContentByName ..
func GetContentByName(name string) (content Content, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var item *badger.Item
		if item, err = txn.Get([]byte(name)); err != nil {
			return
		}
		var data = make([]byte, item.ValueSize())
		if _, err = item.ValueCopy(data); err != nil {
			return
		}
		if err = json.Unmarshal(data, &content); err != nil {
			return
		}
		return
	}); err != nil {
		return
	}
	return
}

// GetContentByURL ..
func GetContentByURL(url string) (content Content, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			var item = iter.Item()
			var data = make([]byte, item.ValueSize())
			if _, err = item.ValueCopy(data); err != nil {
				return
			}
			if err = json.Unmarshal(data, &content); err != nil {
				return
			}
			if content.URL == url {
				return
			}
		}
		return
	}); err != nil {
		return
	}
	err = badger.ErrKeyNotFound
	return
}

// GetContentsByStatus ..
func GetContentsByStatus(status Status) (contents []Content, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			var item = iter.Item()
			var data = make([]byte, item.ValueSize())
			if _, err = item.ValueCopy(data); err != nil {
				return
			}
			var content Content
			if err = json.Unmarshal(data, &content); err != nil {
				return
			}
			if content.Status == status {
				contents = append(contents, content)
			}
		}
		return
	}); err != nil {
		return
	}
	if len(contents) == 0 {
		err = badger.ErrKeyNotFound
	}
	return
}

// GetContents get all of the content
func GetContents() (contents []Content, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			var item = iter.Item()
			var data = make([]byte, item.ValueSize())
			if _, err = item.ValueCopy(data); err != nil {
				return
			}
			var content Content
			if err = json.Unmarshal(data, &content); err != nil {
				return
			}
			contents = append(contents, content)
		}
		return
	}); err != nil {
		return
	}
	if len(contents) == 0 {
		err = badger.ErrKeyNotFound
	}
	return
}

// GenName generate a unique name
func GenName() (name string, err error) {
	var hash = fnv.New64()
again:
	if _, err = hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10))); err != nil {
		return
	}
	name = hex.EncodeToString(hash.Sum(nil))
	if _, err = GetContentByName(name); err == badger.ErrKeyNotFound {
		err = nil
		return
	} else if err != nil {
		return
	}
	<-time.After(time.Microsecond * 100)
	goto again
}
