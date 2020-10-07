package main

import (
	"encoding/hex"
	"hash/fnv"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Status string

const (
	DoneStatus  Status = "done"
	ErrorStatus Status = "error"
	DoingStatus Status = "doing"
)

// Task ..
type Task struct {
	gorm.Model
	Name     string
	URL      string
	Filename string
	Status   Status
}

var DBEngine *gorm.DB

// Database ..
func Database() (err error) {
	if DBEngine, err = gorm.Open(sqlite.Open("transfer.db"), &gorm.Config{}); err != nil {
		return
	}
	if err = DBEngine.AutoMigrate(&Task{}); err != nil {
		return
	}
	return
}

// GenName ..
func GenName() (name string, err error) {
	var hash = fnv.New64()
again:
	if _, err = hash.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10))); err != nil {
		return
	}
	name = hex.EncodeToString(hash.Sum(nil))
	var task Task
	if err = DBEngine.Where(&Task{Name: name}).First(&task).Error; err == gorm.ErrRecordNotFound {
		err = nil
		return
	} else if err != nil {
		return
	} else {
		<-time.After(time.Microsecond * 100)
		goto again
	}
}

// Insert ..
func (t *Task) Insert() (err error) {
	if err = DBEngine.Create(t).Error; err != nil {
		return
	}
	return
}

// UpdateStatus ..
func (t *Task) UpdateStatus() (err error) {
	if err = DBEngine.Debug().Model(&Task{}).Where(t.ID).Updates(t).Error; err != nil {
		return
	}
	return
}
