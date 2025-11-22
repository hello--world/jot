package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// GetFileCreationTime 获取文件的创建时间
// Windows: 通过 syscall 获取真实的创建时间
// Linux/Unix: 使用修改时间作为近似值（大多数文件系统不存储创建时间）
// 其他平台: 使用修改时间作为近似值
func GetFileCreationTime(path string) (time.Time, error) {
	// 获取文件信息
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}

	// Windows: 通过 Win32FileAttributeData 获取创建时间
	// 使用 build tags 来避免在非 Windows 平台上编译错误
	if runtime.GOOS == "windows" {
		creationTime := GetFileCreationTimeWindows(info)
		if !creationTime.IsZero() {
			return creationTime, nil
		}
	}

	// Linux/Unix: 标准库的 Stat_t 不包含 birthtime 字段
	// 虽然 ext4 等文件系统支持创建时间，但需要通过 statx 系统调用获取
	// 这需要 CGO 或额外的系统调用，为了简化，这里使用修改时间作为近似值
	// 注意：在 Linux 上，大多数文件系统不存储创建时间，只有修改时间和访问时间
	// 如果文件系统支持，可以通过 statx 系统调用获取，但需要额外的实现
	// 这里统一使用修改时间作为回退

	// 其他平台或获取失败时，使用修改时间作为近似值
	return info.ModTime(), nil
}

// ParseFileSize parses a file size string like "10M", "100MB", "1G" into bytes
func ParseFileSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Remove "B" or "BYTE" suffix if present (case insensitive)
	sizeStr = strings.ToUpper(sizeStr)
	sizeStr = strings.TrimSuffix(sizeStr, "BYTES")
	sizeStr = strings.TrimSuffix(sizeStr, "BYTE")
	sizeStr = strings.TrimSuffix(sizeStr, "B")
	sizeStr = strings.TrimSpace(sizeStr)

	if sizeStr == "" {
		return 0, fmt.Errorf("invalid size format")
	}

	// Find the last non-digit character to determine the unit
	var numStr string
	var unit string
	for i := len(sizeStr) - 1; i >= 0; i-- {
		if sizeStr[i] >= '0' && sizeStr[i] <= '9' || sizeStr[i] == '.' {
			numStr = sizeStr[:i+1]
			unit = sizeStr[i+1:]
			break
		}
	}
	if numStr == "" {
		numStr = sizeStr
		unit = ""
	}

	// Parse the number
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number: %v", err)
	}

	// Convert based on unit
	unit = strings.ToUpper(strings.TrimSpace(unit))
	switch unit {
	case "", "B":
		return int64(num), nil
	case "K", "KB":
		return int64(num * 1024), nil
	case "M", "MB":
		return int64(num * 1024 * 1024), nil
	case "G", "GB":
		return int64(num * 1024 * 1024 * 1024), nil
	case "T", "TB":
		return int64(num * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// GetTotalFileSize calculates the total size of all files in savePath and uploadPath (excluding backupPath)
func GetTotalFileSize(savePath, uploadPath string) (int64, error) {
	var totalSize int64

	// Calculate size of notes in savePath
	if err := filepath.Walk(savePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	// Calculate size of uploaded files in uploadPath
	if err := filepath.Walk(uploadPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	}); err != nil && !os.IsNotExist(err) {
		return 0, err
	}

	return totalSize, nil
}

// LoadEnvFile 从 .env 文件加载环境变量
func LoadEnvFile() error {
	file, err := os.Open(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil // .env file is optional
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}
