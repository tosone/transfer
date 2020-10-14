package v2

import (
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/tosone/logging"
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
)

type Content struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	URL            string `json:"url"`
	Filename       string `json:"filename"`
	RandomFilename bool   `json:"randomFilename"`
	Path           string `json:"path"`
	Bucket         string `json:"bucket"`
	Region         string `json:"region"`
	Endpoint       string `json:"endpoint"`
	Force          bool   `json:"force"`

	Progress string `json:"progress"`

	Status  Status `json:"-"`
	Message string `json:"-"`
	Content []byte `json:"-"`
}

var DBEngine *badger.DB

const databaseFile = "badgerDir"

func init() {
	var err error
	if DBEngine, err = badger.Open(badger.DefaultOptions(databaseFile)); err != nil {
		logging.Fatal(err)
	}
}

// Teardown ..
func Teardown() (err error) {
	if err = DBEngine.Close(); err != nil {
		return
	}
	return
}

// Insert ..
func (t *Content) Insert() (err error) {
	var txn = DBEngine.NewTransaction(true)

	var data []byte
	if data, err = json.Marshal(*t); err != nil {
		return
	}

	var entry = badger.NewEntry([]byte(t.Name), data)
	if err = txn.SetEntry(entry); err != nil {
		return
	}
	if err = txn.Commit(); err != nil {
		txn.Discard()
		return
	}
	return
}

// UpdateStatus ..
func (t *Content) UpdateStatus(status Status) (err error) {
	var txn = DBEngine.NewTransaction(true)
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
	if err = DBEngine.View(func(txn *badger.Txn) (err error) {
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
	if err = DBEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
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
