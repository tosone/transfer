package database

import (
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/spf13/viper"
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
	// DoingStatus ..
	DoingStatus Status = "doing"
)

type Task struct {
	Name           string  `json:"name"`
	UploadType     string  `json:"uploadType"`
	URL            string  `json:"url"`
	DownloadType   string  `json:"downloadType"`
	DownloadURL    string  `json:"downloadUrl"` // user can download file from this url
	Filename       string  `json:"filename"`
	RandomFilename bool    `json:"randomFilename"`
	Path           string  `json:"path"`
	Force          bool    `json:"force"`
	Progress       float64 `json:"progress"`
	Status         Status  `json:"status"`
	Message        string  `json:"message"`
}

var dbEngine *badger.DB

// Initialize ..
func Initialize() (err error) {
	var options = badger.DefaultOptions(viper.GetString("DatabaseDir"))
	options.Logger = &logging.Inst{}
	if dbEngine, err = badger.Open(options); err != nil {
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
func (t *Task) Insert() (err error) {
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
func (t *Task) UpdateStatus(status Status) (err error) {
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

// GetTaskByName ..
func GetTaskByName(name string) (task Task, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var item *badger.Item
		if item, err = txn.Get([]byte(name)); err != nil {
			return
		}
		var data = make([]byte, item.ValueSize())
		if _, err = item.ValueCopy(data); err != nil {
			return
		}
		if err = json.Unmarshal(data, &task); err != nil {
			return
		}
		return
	}); err != nil {
		return
	}
	return
}

// GetTaskByURL ..
func GetTaskByURL(url string) (task Task, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			var item = iter.Item()
			var data = make([]byte, item.ValueSize())
			if _, err = item.ValueCopy(data); err != nil {
				return
			}
			if err = json.Unmarshal(data, &task); err != nil {
				return
			}
			if task.URL == url {
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

// GetTasksByStatus ..
func GetTasksByStatus(status Status) (tasks []Task, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			var item = iter.Item()
			var data = make([]byte, item.ValueSize())
			if _, err = item.ValueCopy(data); err != nil {
				return
			}
			var task Task
			if err = json.Unmarshal(data, &task); err != nil {
				return
			}
			if task.Status == status {
				tasks = append(tasks, task)
			}
		}
		return
	}); err != nil {
		return
	}
	if len(tasks) == 0 {
		err = badger.ErrKeyNotFound
	}
	return
}

// GetTasks get all of the content
func GetTasks() (tasks []Task, err error) {
	if err = dbEngine.View(func(txn *badger.Txn) (err error) {
		var iter = txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			var item = iter.Item()
			var data = make([]byte, item.ValueSize())
			if _, err = item.ValueCopy(data); err != nil {
				return
			}
			var task Task
			if err = json.Unmarshal(data, &task); err != nil {
				return
			}
			tasks = append(tasks, task)
		}
		return
	}); err != nil {
		return
	}
	if len(tasks) == 0 {
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
	if _, err = GetTaskByName(name); err == badger.ErrKeyNotFound {
		err = nil
		return
	} else if err != nil {
		return
	}
	<-time.After(time.Microsecond * 100)
	goto again
}
