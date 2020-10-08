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
	DoneStatus    Status = "done"
	ErrorStatus   Status = "error"
	PendingStatus Status = "pending"
)

var DBEngine *gorm.DB

const DatabaseFile = "transfer.db"

// Database ..
func Database() (err error) {
	if DBEngine, err = gorm.Open(sqlite.Open(DatabaseFile), &gorm.Config{}); err != nil {
		return
	}
	if err = DBEngine.AutoMigrate(&Content{}); err != nil {
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
	var task Content
	if err = DBEngine.Where(&Content{Name: name}).First(&task).Error; err == gorm.ErrRecordNotFound {
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
func (t *Content) Insert() (err error) {
	if err = DBEngine.Debug().Create(t).Error; err != nil {
		return
	}
	return
}

// UpdateStatus ..
func (t *Content) UpdateStatus() (err error) {
	if err = DBEngine.Model(&Content{}).Where(t.Name).Updates(t).Error; err != nil {
		return
	}
	return
}
