package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillManager 管理技能的核心逻辑：扫描源目录、检测链接状态、创建/删除链接。
type SkillManager struct {
	sourceDir string
	targets   []TargetInfo
}

// NewSkillManager 创建新的 SkillManager 实例。
func NewSkillManager() *SkillManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("USERPROFILE")
	}

	return &SkillManager{
		sourceDir: filepath.Join(homeDir, ".cc-switch", "skills"),
		targets: []TargetInfo{
			{Key: "opencode", Label: "OpenCode", Path: filepath.Join(homeDir, ".config", "opencode", "skills")},
			{Key: "claude", Label: "Claude Code", Path: filepath.Join(homeDir, ".claude", "skills")},
			{Key: "codex", Label: "Codex", Path: filepath.Join(homeDir, ".codex", "skills")},
		},
	}
}

// GetTargets 返回已配置的目标平台列表。
func (sm *SkillManager) GetTargets() []TargetInfo {
	return sm.targets
}

// GetAllSkills 扫描源目录，返回所有技能及其在各平台的链接状态。
func (sm *SkillManager) GetAllSkills() []SkillInfo {
	entries, err := os.ReadDir(sm.sourceDir)
	if err != nil {
		return nil
	}

	skills := make([]SkillInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// 跳过隐藏目录
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		skillPath := filepath.Join(sm.sourceDir, entry.Name())
		meta := parseSkillMeta(skillPath)

		// 始终以目录名作为技能主键（Name），保证文件系统操作正确
		// SKILL.md 的 frontmatter name 可能和目录名不一致（如 afsim vs afsim-scripts）
		displayName := meta.Name
		if displayName == "" || displayName == entry.Name() {
			displayName = entry.Name()
		}

		si := SkillInfo{
			Name:        entry.Name(), // 目录名 = 文件系统主键
			Description: meta.Description,
			SourcePath:  skillPath,
			Targets:     make(map[string]bool),
		}

		// 如果 frontmatter 名与目录名不同，在描述前追加显示名
		if meta.Name != "" && meta.Name != entry.Name() {
			if si.Description != "" {
				si.Description = "[" + meta.Name + "] " + si.Description
			} else {
				si.Description = "[" + meta.Name + "]"
			}
		}

		// 检测各平台的链接状态
		for _, t := range sm.targets {
			si.Targets[t.Key] = isLinked(skillPath, t.Path)
		}

		skills = append(skills, si)
	}

	return skills
}

// skillMeta 从 SKILL.md 解析出的元数据。
type skillMeta struct {
	Name        string
	Description string
}

// parseSkillMeta 解析技能目录中 SKILL.md 的 YAML frontmatter。
func parseSkillMeta(skillPath string) skillMeta {
	skillMd := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(skillMd)
	if err != nil {
		return skillMeta{}
	}

	content := string(data)
	return skillMeta{
		Name:        extractFrontmatterField(content, "name"),
		Description: extractFrontmatterField(content, "description"),
	}
}

// extractFrontmatterField 从 YAML frontmatter 中提取指定字段的值。
// 简单解析，避免引入完整的 YAML 解析库依赖。
func extractFrontmatterField(content, field string) string {
	// 检查是否以 --- 开头（frontmatter 标记）
	if !strings.HasPrefix(content, "---") {
		return ""
	}

	// 找到第二个 --- 的位置
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				break // frontmatter 结束
			}
		}
		if inFrontmatter {
			// 解析 "field: value" 格式
			if strings.HasPrefix(trimmed, field+":") {
				value := strings.TrimSpace(strings.TrimPrefix(trimmed, field+":"))
				// 去掉可能的引号
				value = strings.Trim(value, "\"'")
				return value
			}
		}
	}

	return ""
}

// ========== 技能链接操作 ==========

// ToggleSkill 切换某个技能在指定目标的链接状态。
func (sm *SkillManager) ToggleSkill(skillName, target string, enable bool) (bool, error) {
	// 查找目标平台配置
	var targetPath string
	for _, t := range sm.targets {
		if t.Key == target {
			targetPath = t.Path
			break
		}
	}
	if targetPath == "" {
		return false, fmt.Errorf("未知的目标平台: %s", target)
	}

	// 确保目标目录存在
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return false, fmt.Errorf("无法创建目标目录 %s: %w", targetPath, err)
	}

	skillPath := filepath.Join(sm.sourceDir, skillName)
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return false, fmt.Errorf("技能目录不存在: %s", skillPath)
	}

	linkPath := filepath.Join(targetPath, skillName)

	if enable {
		// 如果已经链接，先移除再创建
		if isLinked(skillPath, targetPath) {
			if err := removeLink(linkPath); err != nil {
				return false, fmt.Errorf("移除已有链接失败: %w", err)
			}
		}
		if err := createLink(skillPath, linkPath); err != nil {
			return false, fmt.Errorf("创建链接失败: %w", err)
		}
		return true, nil
	} else {
		if err := removeLink(linkPath); err != nil {
			// 如果链接本就不存在，不算错误
			if os.IsNotExist(err) {
				return false, nil
			}
			return false, fmt.Errorf("移除链接失败: %w", err)
		}
		return false, nil
	}
}

// ToggleAll 批量切换某个目标平台下所有技能的链接状态。
func (sm *SkillManager) ToggleAll(target string, enable bool) []error {
	skills := sm.GetAllSkills()
	errs := make([]error, 0)

	for _, sk := range skills {
		_, err := sm.ToggleSkill(sk.Name, target, enable)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
