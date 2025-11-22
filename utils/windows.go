//go:build windows
// +build windows

package utils

import (
	"os"
	"syscall"
	"time"
)

// GetFileCreationTimeWindows 在 Windows 平台上获取文件创建时间
func GetFileCreationTimeWindows(info os.FileInfo) time.Time {
	if sys, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		// Windows 文件创建时间是从 1601-01-01 00:00:00 UTC 开始的 100 纳秒间隔
		// 转换为 Unix 时间
		ft := sys.CreationTime
		// 合并高32位和低32位值（使用 uint64 避免溢出）
		nsec100 := uint64(ft.HighDateTime)<<32 | uint64(ft.LowDateTime)
		// Windows 纪元: 1601-01-01 00:00:00 UTC = 116444736000000000 (100纳秒间隔)
		const windowsEpoch100ns = uint64(116444736000000000)
		// 减去 Windows 纪元并转换为纳秒
		unixNsec := int64((nsec100 - windowsEpoch100ns) * 100)
		creationTime := time.Unix(0, unixNsec)
		return creationTime
	}
	return time.Time{}
}

