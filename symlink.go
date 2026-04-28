package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// createLink 创建从 source 到 dest 的目录符号链接。
// Windows 上优先使用 os.Symlink（需要开发者模式），
// 失败时回退到 mklink /J（目录联接，无需管理员权限）。
func createLink(source, dest string) error {
	// 确保目标父目录存在
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 先尝试标准符号链接
	err := os.Symlink(source, dest)
	if err == nil {
		return nil
	}

	// Windows 回退：使用 mklink /J 创建目录联接
	if runtime.GOOS == "windows" {
		return createJunction(source, dest)
	}

	return fmt.Errorf("创建符号链接失败: %w", err)
}

// removeLink 安全地移除符号链接或目录联接。
// 只移除链接本身，不会删除源目录内容。
func removeLink(linkPath string) error {
	info, err := os.Lstat(linkPath)
	if err != nil {
		return err
	}

	// 符号链接：直接删除
	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(linkPath)
	}

	// 目录联接：Windows 上需要用特殊方式删除
	if runtime.GOOS == "windows" && isReparsePoint(linkPath) {
		return removeJunction(linkPath)
	}

	// 普通目录不应该通过此函数删除，防止误删
	return fmt.Errorf("路径存在但不是链接，拒绝删除: %s", linkPath)
}

// createJunction 在 Windows 上创建目录联接（Junction）。
// 目录联接不要求管理员权限，且对大多数程序透明。
func createJunction(source, dest string) error {
	// mklink /J 在 cmd 中运行，需要绝对路径
	absSource, err := filepath.Abs(source)
	if err != nil {
		return fmt.Errorf("获取源目录绝对路径失败: %w", err)
	}
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("获取目标路径失败: %w", err)
	}

	// 如果目标已存在（且是联接点），先删除
	if isReparsePoint(absDest) {
		if err := removeJunction(absDest); err != nil {
			return err
		}
	}

	cmd := exec.Command("cmd", "/c", "mklink", "/J", absDest, absSource)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink /J 失败: %w\n输出: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

// removeJunction 在 Windows 上删除目录联接（Junction）。
// 使用 rmdir 而非 rd /s，因为 /s 会删除目录内容。
func removeJunction(junctionPath string) error {
	cmd := exec.Command("cmd", "/c", "rmdir", junctionPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rmdir 失败: %w\n输出: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// isReparsePoint 检测路径是否为 Windows 重解析点（Reparse Point）。
// Go 1.23+ 的 os.Readlink 在 Windows 上通过 FSCTL_GET_REPARSE_POINT 实现，
// 能正确识别符号链接和目录联接（Junction）。
func isReparsePoint(path string) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	_, err := os.Readlink(path)
	return err == nil
}

// isLinked 检查技能是否在目标目录中存在有效的链接（符号链接或目录联接）。
// 本函数在 skills.go 中通过 isReparsePoint 间接调用，此处提供备选实现。
func isLinked(skillPath, targetDir string) bool {
	skillName := filepath.Base(skillPath)
	linkPath := filepath.Join(targetDir, skillName)

	info, err := os.Lstat(linkPath)
	if err != nil {
		return false
	}

	// 符号链接
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}

	// Windows 目录联接
	if runtime.GOOS == "windows" && info.IsDir() {
		return isReparsePoint(linkPath)
	}

	return false
}
