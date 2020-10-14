package v2

import (
	"fmt"
	"testing"

	"github.com/dgraph-io/badger/v2"
)

func TestContent_Insert(t *testing.T) {
	var err error
	var content = &Content{URL: "test", Name: "123"}
	if err = content.Insert(); err != nil {
		t.Error(err)
	}
	if err = content.UpdateStatus(DoneStatus); err != nil {
		t.Error(err)
	}
	if _, err = GetContentByName("123"); err != nil {
		t.Error(err)
	}
	if _, err = GetContentByName("test"); err != badger.ErrKeyNotFound {
		t.Error(err)
	}
	var name string
	if name, err = GenName(); err != nil {
		t.Error(err)
	}
	fmt.Printf("name: %s\n", name)
	if err = Teardown(); err != nil {
		t.Error(err)
	}
}
