package router

import (
	"fmt"
	"sync"

	"transfer/counter"
)

// ProgressBarPool ..
var ProgressBarPool = sync.Map{}

func getProgress(name string) (progress string) {
	ProgressBarPool.Range(func(key, value interface{}) bool {
		if key.(string) == name {
			var reader = value.(*counter.Reader)
			progress = fmt.Sprintf("%.2f", reader.Percent)
			return false
		}
		return true
	})
	return
}
