package router

import (
	"fmt"
	"strconv"
	"sync"

	"transfer/counter"
)

// ProgressBarPool ..
var ProgressBarPool = sync.Map{}

func getProgress(name string) (progress float64, err error) {
	ProgressBarPool.Range(func(key, value interface{}) bool {
		if key.(string) == name {
			var reader = value.(*counter.Reader)
			if progress, err = strconv.ParseFloat(fmt.Sprintf("%.2f", reader.Percent), 64); err != nil {
				return false
			}
			return false
		}
		return true
	})
	return
}
