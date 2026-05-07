// Package skill 管理 opencode 技能：扫描技能目录、检测链接状态、读写 SKILL.md。
package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"oc-manager/model"
)

// Manager 管理 opencode 技能。
type Manager struct {
	globalDir string // ~/.config/opencode/skills/
}

// NewManager 创建新的 Manager 实例。
func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return &Manager{
		globalDir: filepath.Join(homeDir, ".config", "opencode", "skills"),
	}
}

// SourceDir 返回全局技能目录路径。
func (m *Manager) SourceDir() string {
	return m.globalDir
}

// skillFrontmatter SKILL.md frontmatter 解析结果。
type skillFrontmatter struct {
	Name        string
	Description string
}

// GetAllSkills 扫描全局技能目录，返回所有技能信息。
func (m *Manager) GetAllSkills() []model.SkillInfo {
	return m.scanDir(m.globalDir, "global")
}

// scanDir 扫描指定目录下的技能子目录。
func (m *Manager) scanDir(dir, source string) []model.SkillInfo {
	var skills []model.SkillInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return skills
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(dir, skillName)
		skillMD := filepath.Join(skillPath, "SKILL.md")

		// 解析 SKILL.md 获取 name 和 description
		name := skillName
		desc := ""
		if data, err := os.ReadFile(skillMD); err == nil {
			fm, _ := parseFrontmatter(data)
			if fm.Name != "" {
				name = fm.Name
			}
			desc = fm.Description
		}

		// 检测链接状态
		linked := false
		linkPath := filepath.Join(m.globalDir, skillName)
		if info, err := os.Lstat(linkPath); err == nil {
			linked = (info.Mode()&os.ModeSymlink != 0) || info.IsDir()
		}

		skills = append(skills, model.SkillInfo{
			Name:        name,
			Description: desc,
			Path:        skillPath,
			Linked:      linked,
			Source:      source,
		})
	}

	return skills
}

// ReadSkillContent 读取指定技能目录下的 SKILL.md 完整内容。
func (m *Manager) ReadSkillContent(skillPath string) (string, error) {
	mdPath := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		return "", fmt.Errorf("读取 SKILL.md 失败: %w", err)
	}
	return string(data), nil
}

// SaveSkillContent 保存 SKILL.md 内容到指定技能目录。
func (m *Manager) SaveSkillContent(skillPath, content string) error {
	mdPath := filepath.Join(skillPath, "SKILL.md")
	return os.WriteFile(mdPath, []byte(content), 0644)
}

// ToggleSkill 切换技能链接状态。
func (m *Manager) ToggleSkill(skillPath, skillName string, enable bool) (bool, error) {
	linkPath := filepath.Join(m.globalDir, skillName)

	if enable {
		if runtime.GOOS == "windows" {
			cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, skillPath)
			if err := cmd.Run(); err != nil {
				if err := os.Symlink(skillPath, linkPath); err != nil {
					return false, fmt.Errorf("创建符号链接失败: %w", err)
				}
			}
		} else {
			if err := os.Symlink(skillPath, linkPath); err != nil {
				return false, fmt.Errorf("创建符号链接失败: %w", err)
			}
		}
		return true, nil
	} else {
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("删除符号链接失败: %w", err)
		}
		return false, nil
	}
}

// parseFrontmatter 手动解析 Markdown YAML frontmatter。
func parseFrontmatter(data []byte) (skillFrontmatter, error) {
	content := string(data)
	if len(content) < 4 || content[:3] != "---" {
		return skillFrontmatter{}, fmt.Errorf("无 frontmatter")
	}
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		// 尝试 "---" 紧跟 rest
		end = strings.Index(rest, "---")
	}
	if end == -1 {
		return skillFrontmatter{}, fmt.Errorf("frontmatter 未闭合")
	}
	fmText := rest[:end]
	var fm skillFrontmatter
	for _, line := range strings.Split(fmText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"`)
		switch key {
		case "name":
			fm.Name = val
		case "description":
			fm.Description = val
		}
	}
	return fm, nil
}
