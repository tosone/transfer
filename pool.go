package main

import (
	"sync"
)

// DownloadPool ..
var DownloadPool = sync.Map{}

// DownloadTask ..
//type DownloadTask struct {
//	Content     Content         `json:"content"`
//	Progress    string          `json:"progress"`
//	ProgressBar *pb.ProgressBar `json:"-"`
//}

// ProgressBarPool ..
var ProgressBarPool = sync.Map{}
