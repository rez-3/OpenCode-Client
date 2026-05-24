// Package fileutil 提供原子文件写入和 JSONC 验证工具。
package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AtomicWrite 将数据原子写入 path（临时文件 + fsync + 重命名），避免文件损坏。
// 写入前会验证内容是否为有效 JSON/JSONC。
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	if strings.TrimSpace(string(data)) == "" {
		return fmt.Errorf("拒绝写入空配置文件: %s", path)
	}
	if err := ValidateJSONC(data); err != nil {
		return fmt.Errorf("拒绝写入无效配置文件 %s: %w", path, err)
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时配置文件失败: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("写入临时配置文件失败: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("设置临时配置文件权限失败: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("同步临时配置文件失败: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭临时配置文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("替换配置文件失败: %w", err)
	}
	return nil
}

// ValidateJSONC 验证 JSONC 数据是否包含有效的 JSON（去除注释后）。
func ValidateJSONC(data []byte) error {
	cleaned := strings.TrimSpace(StripComments(string(data)))
	if cleaned == "" {
		return fmt.Errorf("配置内容为空")
	}
	if !json.Valid([]byte(cleaned)) {
		return fmt.Errorf("配置内容不是有效 JSON/JSONC")
	}
	return nil
}

// StripComments 移除 JSONC 中的单行注释（// ...），正确处理引号内的 //。
func StripComments(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		idx := strings.Index(line, "//")
		if idx >= 0 {
			before := line[:idx]
			if strings.Count(before, "\"")%2 == 0 {
				line = before
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}
