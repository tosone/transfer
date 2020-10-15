package router

import (
	"fmt"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

// ProgressBarPool ..
var ProgressBarPool = sync.Map{}

func getProgress(name string) (progress string) {
	ProgressBarPool.Range(func(key, value interface{}) bool {
		if key.(string) == name {
			var bar = value.(*pb.ProgressBar)
			progress = fmt.Sprintf("%.2f", float64(bar.Current()*100.0)/float64(bar.Total()))
			return false
		}
		return true
	})
	return
}
