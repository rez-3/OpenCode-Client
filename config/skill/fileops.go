package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"oc-manager/model"
)

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
	root, err := m.resolveSkillPath(skillPath)
	if err != nil {
		return err
	}
	mdPath := filepath.Join(root, "SKILL.md")
	return os.WriteFile(mdPath, []byte(content), 0644)
}

// ListSkillFiles 返回技能根目录下的只读文件树。
func (m *Manager) ListSkillFiles(skillPath string) (model.SkillFileNode, error) {
	root, err := m.resolveSkillPath(skillPath)
	if err != nil {
		return model.SkillFileNode{}, err
	}
	node, err := m.buildSkillFileTree(root, root, true)
	if err != nil {
		return model.SkillFileNode{}, err
	}
	node.Name = filepath.Base(filepath.Clean(skillPath))
	return node, nil
}

// ReadSkillFile 读取技能根目录内的文本文件。
func (m *Manager) ReadSkillFile(skillPath, relativePath string) (model.SkillContent, error) {
	root, err := m.resolveSkillPath(skillPath)
	if err != nil {
		return model.SkillContent{}, err
	}
	target, rel, err := m.resolveSkillRelativePath(root, relativePath)
	if err != nil {
		return model.SkillContent{}, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return model.SkillContent{}, fmt.Errorf("读取技能文件失败: %w", err)
	}
	if info.IsDir() {
		return model.SkillContent{}, fmt.Errorf("目标不是文本文件")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return model.SkillContent{}, fmt.Errorf("读取技能文件失败: %w", err)
	}
	if !isTextContent(data) {
		return model.SkillContent{}, fmt.Errorf("仅支持预览文本文件")
	}
	return model.SkillContent{Path: rel, Content: string(data)}, nil
}

// SaveSkillFile 保存技能根目录内的文本文件。
func (m *Manager) SaveSkillFile(skillPath, relativePath, content string) error {
	root, err := m.resolveSkillPath(skillPath)
	if err != nil {
		return err
	}
	target, _, err := m.resolveSkillRelativePath(root, relativePath)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("读取技能文件失败: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("目标不是文本文件")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("读取技能文件失败: %w", err)
	}
	if !isTextContent(data) {
		return fmt.Errorf("仅支持编辑文本文件")
	}
	return os.WriteFile(target, []byte(content), info.Mode().Perm())
}
