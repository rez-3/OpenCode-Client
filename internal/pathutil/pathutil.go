// Package pathutil 提供跨平台的路径规范化与比较工具。
package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Normalize 将路径规范化为绝对路径并清理分隔符。
// Windows 上统一转为小写以实现大小写不敏感比较。
func Normalize(p string) (string, error) {
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	cleaned := filepath.Clean(absPath)
	if runtime.GOOS == "windows" {
		cleaned = strings.ToLower(cleaned)
	}
	return cleaned, nil
}

// Equal 比较两个路径是否指向同一位置。
func Equal(a, b string) bool {
	na, err := Normalize(a)
	if err != nil {
		return false
	}
	nb, err := Normalize(b)
	if err != nil {
		return false
	}
	return na == nb
}

// CompareKey 返回用于排序比较的规范化路径键值。
func CompareKey(p string) string {
	normalized, err := Normalize(p)
	if err != nil {
		return strings.ToLower(filepath.Clean(p))
	}
	return normalized
}

// IsSubPath 检查 child 是否在 parent 目录之下。
func IsSubPath(child, parent string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, "..")
}

// DirExists 检查指定路径是否为存在的目录。
func DirExists(p string) bool {
	if runtime.GOOS == "windows" {
		// Windows 上盘符根目录（如 C:\）始终存在
		if strings.HasSuffix(filepath.Clean(p), ":") {
			return true
		}
	}
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
