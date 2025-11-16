//go:build !windows
// +build !windows

package main

import (
	"os"
	"time"
)

// getFileCreationTimeWindows 在非 Windows 平台上的 stub 实现
// 返回零值，调用者会使用修改时间作为回退
func getFileCreationTimeWindows(info os.FileInfo) time.Time {
	return time.Time{}
}
