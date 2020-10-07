package main

import (
	"sync"

	"github.com/cheggaaa/pb/v3"
)

// DownloadPool ..
var DownloadPool = sync.Map{}

// DownloadTask ..
type DownloadTask struct {
	URL         string          `json:"url"`
	Filename    string          `json:"filename"`
	Progress    string          `json:"progress"`
	ProgressBar *pb.ProgressBar `json:"-"`
}
