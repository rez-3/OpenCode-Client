package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// IsLinked 检查指定技能名称是否已在全局目录中存在链接。
func (m *Manager) IsLinked(skillName string) bool {
	linkPath := filepath.Join(m.globalDir, skillName)
	if info, err := os.Lstat(linkPath); err == nil {
		return (info.Mode()&os.ModeSymlink != 0) || info.IsDir()
	}
	return false
}

// ToggleSkill 切换技能链接状态。
func (m *Manager) ToggleSkill(skillPath, skillName string, enable bool) (bool, error) {
	topName, topSource := resolveTopLevelLink(skillPath, skillName)
	linkPath := filepath.Join(m.globalDir, topName)

	if enable {
		if _, err := os.Readlink(linkPath); err == nil {
			if runtime.GOOS == "windows" {
				rmdirCmd := exec.Command("cmd", "/c", "rmdir", linkPath)
				rmdirCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				rmdirCmd.Run()
			} else {
				os.Remove(linkPath)
			}
		}
		if err := os.Symlink(topSource, linkPath); err == nil {
			return true, nil
		}
		if runtime.GOOS == "windows" {
			absSource, _ := filepath.Abs(topSource)
			absDest, _ := filepath.Abs(linkPath)
			cmd := exec.Command("cmd", "/c", "mklink", "/J", absDest, absSource)
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
			if output, err := cmd.CombinedOutput(); err != nil {
				return false, fmt.Errorf("创建链接失败: %w\n输出: %s", err, strings.TrimSpace(string(output)))
			}
			return true, nil
		}
		return false, fmt.Errorf("创建符号链接失败")
	} else {
		if runtime.GOOS == "windows" {
			if _, err := os.Readlink(linkPath); err == nil {
				rmCmd := exec.Command("cmd", "/c", "rmdir", linkPath)
				rmCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				if output, err := rmCmd.CombinedOutput(); err != nil {
					return false, fmt.Errorf("删除链接失败: %w\n输出: %s", err, strings.TrimSpace(string(output)))
				}
				return false, nil
			}
		}
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("删除链接失败: %w", err)
		}
		return false, nil
	}
}

// resolveTopLevelLink 解析嵌套技能的顶层链接信息。
func resolveTopLevelLink(skillPath, skillName string) (topName, topSource string) {
	osName := filepath.FromSlash(skillName)
	depth := strings.Count(osName, string(filepath.Separator))
	if depth == 0 {
		return skillName, skillPath
	}
	topName = strings.SplitN(skillName, "/", 2)[0]
	topSource = skillPath
	for i := 0; i < depth; i++ {
		topSource = filepath.Dir(topSource)
	}
	return topName, topSource
}

// LinkSkill 创建单个技能链接到全局目录。
func (m *Manager) LinkSkill(skillPath, skillName string) error {
	topName, topSource := resolveTopLevelLink(skillPath, skillName)
	linkPath := filepath.Join(m.globalDir, topName)
	if _, err := os.Lstat(linkPath); err == nil {
		if runtime.GOOS == "windows" {
			if _, err := os.Readlink(linkPath); err == nil {
				rmCmd := exec.Command("cmd", "/c", "rmdir", linkPath)
				rmCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				rmCmd.Run()
			} else {
				os.Remove(linkPath)
			}
		}
	}
	if err := os.Symlink(topSource, linkPath); err == nil {
		return nil
	}
	if runtime.GOOS == "windows" {
		absSource, _ := filepath.Abs(topSource)
		absDest, _ := filepath.Abs(linkPath)
		cmd := exec.Command("cmd", "/c", "mklink", "/J", absDest, absSource)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("创建链接失败: %w\n输出: %s", err, strings.TrimSpace(string(output)))
		}
		return nil
	}
	return fmt.Errorf("创建链接失败")
}

// normalizeComparePath 规范化路径用于前缀比较。
func normalizeComparePath(p string) string {
	cleaned := filepath.Clean(p)
	cleaned = strings.TrimPrefix(cleaned, `\\?\`)
	if runtime.GOOS == "windows" {
		cleaned = strings.ToLower(cleaned)
	}
	return cleaned
}

// isSubPath 判断 child 是否在 parent 目录下（或等于 parent）。
func isSubPath(child, parent string) bool {
	if !strings.HasPrefix(child, parent) {
		return false
	}
	if len(child) == len(parent) {
		return true
	}
	sep := child[len(parent)]
	return sep == filepath.Separator || sep == '/'
}

// GetManagedLinks 获取全局目录中所有被本应用托管的链接。
func (m *Manager) GetManagedLinks(sourceDirs []string) ([]string, error) {
	if len(sourceDirs) == 0 {
		return nil, nil
	}
	var normalizedSourceDirs []string
	for _, d := range sourceDirs {
		absD, err := filepath.Abs(filepath.Clean(d))
		if err != nil {
			continue
		}
		normalizedSourceDirs = append(normalizedSourceDirs, normalizeComparePath(absD))
	}

	entries, err := os.ReadDir(m.globalDir)
	if err != nil {
		return nil, err
	}

	var managed []string
	for _, entry := range entries {
		entryPath := filepath.Join(m.globalDir, entry.Name())
		var target string
		resolved := false

		if linkTarget, err := os.Readlink(entryPath); err == nil && linkTarget != "" {
			if filepath.IsAbs(linkTarget) {
				target = linkTarget
			} else {
				target = filepath.Join(m.globalDir, linkTarget)
			}
			resolved = true
		}
		if !resolved {
			if t, err := filepath.EvalSymlinks(entryPath); err == nil && t != entryPath {
				target = t
				resolved = true
			}
		}
		if !resolved {
			continue
		}

		target = normalizeComparePath(target)
		for _, src := range normalizedSourceDirs {
			if isSubPath(target, src) {
				managed = append(managed, entry.Name())
				break
			}
		}
	}
	return managed, nil
}

// ClearManagedLinks 删除全局目录中的所有托管链接。
func (m *Manager) ClearManagedLinks(sourceDirs []string) error {
	managed, err := m.GetManagedLinks(sourceDirs)
	if err != nil {
		return err
	}
	for _, name := range managed {
		linkPath := filepath.Join(m.globalDir, name)
		os.Remove(linkPath)
	}
	return nil
}
