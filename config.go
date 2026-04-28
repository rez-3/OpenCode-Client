package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ========== 配置数据结构 ==========

// OpenAgentConfig 表示 oh-my-openagent.jsonc 的顶层结构。
type OpenAgentConfig struct {
	Agents     map[string]ModelConfig `json:"agents"`
	Categories map[string]ModelConfig `json:"categories"`
}

// ModelConfig 单个 agent/category 的模型配置。
type ModelConfig struct {
	Model string `json:"model"`
}

// ModelEntry 前端展示用的模型条目（agent 或 category）。
type ModelEntry struct {
	Key      string `json:"key"`      // agent/category 名称
	Type     string `json:"type"`     // "agent" 或 "category"
	Model    string `json:"model"`    // 当前模型
	Label    string `json:"label"`    // 中文简称（从注释提取，≤5字）
	Comment  string `json:"comment"`  // 原始注释文本
}

// ========== 配置路径 & 加载 ==========

// configPath 返回 oh-my-openagent.jsonc 的完整路径。
func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "opencode", "oh-my-openagent.jsonc")
}

// loadConfig 读取并解析 JSONC 配置，同时返回原始文本用于后续写回。
func loadConfig() (*OpenAgentConfig, string, map[string]string, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	rawText := string(data)

	// 去掉单行注释后解析 JSON
	cleaned := stripComments(rawText)

	var config OpenAgentConfig
	if err := json.Unmarshal([]byte(cleaned), &config); err != nil {
		return nil, rawText, nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 提取注释（key → comment text）
	comments := extractComments(rawText)

	return &config, rawText, comments, nil
}

// saveConfig 将模型变更写回 JSONC 文件，保留原有格式和注释。
// 使用行级匹配替换，保证非模型字段和注释不受影响。
func saveConfig(entries []ModelEntry) error {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	modelRe := regexp.MustCompile(`("model"\s*:\s*)"[^"]*"`)

	for i, line := range lines {
		for _, entry := range entries {
			// 查找包含当前 entry key 的行（如 "oracle": { 开头）
			trimmed := strings.TrimSpace(line)
			keyPattern := fmt.Sprintf(`"%s":`, entry.Key)

			// 标记是否进入目标 block
			if strings.Contains(trimmed, keyPattern) && strings.Contains(trimmed, "{") {
				// model 可能在同一行
				if modelRe.MatchString(line) {
					lines[i] = modelRe.ReplaceAllString(line, fmt.Sprintf(`${1}"%s"`, entry.Model))
					break
				}
				// 也可能在后续行
				for j := i + 1; j < len(lines) && j < i+5; j++ {
					if modelRe.MatchString(lines[j]) {
						lines[j] = modelRe.ReplaceAllString(lines[j], fmt.Sprintf(`${1}"%s"`, entry.Model))
						break
					}
					// 遇到 } 则 block 结束
					if strings.Contains(lines[j], "}") {
						break
					}
				}
				break
			}
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// stripComments 移除 JSONC 中的单行注释（// ...）。
func stripComments(text string) string {
	lines := strings.Split(text, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		// 只在双引号外的 // 才算注释（简单处理，实际场景中 model 值不含 //）
		idx := strings.Index(line, "//")
		if idx >= 0 {
			// 确保 // 不在引号内（简化：检查 // 前引号数量）
			before := line[:idx]
			if strings.Count(before, "\"")%2 == 0 {
				line = before
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// ========== 获取模型列表 ==========

// getAvailableModels 执行 opencode models 获取可用模型列表。
func getAvailableModels() ([]string, error) {
	return runModelsCmd()
}

// ========== 配置转前端结构 ==========

// configToEntries 将 OpenAgentConfig 转为前端展示用的 ModelEntry 列表。
func configToEntries(config *OpenAgentConfig, comments map[string]string) []ModelEntry {
	entries := make([]ModelEntry, 0)

	// Agents
	for key, mc := range config.Agents {
		comment := comments[key]
		entries = append(entries, ModelEntry{
			Key:     key,
			Type:    "agent",
			Model:   mc.Model,
			Label:   deriveLabel(comment),
			Comment: comment,
		})
	}

	// Categories
	for key, mc := range config.Categories {
		comment := comments[key]
		entries = append(entries, ModelEntry{
			Key:     key,
			Type:    "category",
			Model:   mc.Model,
			Label:   deriveLabel(comment),
			Comment: comment,
		})
	}

	return entries
}

// ========== 注释提取 ==========

// extractComments 从原始 JSONC 文本中提取每个 key 对应的行内注释。
func extractComments(rawText string) map[string]string {
	lines := strings.Split(rawText, "\n")
	comments := make(map[string]string)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 匹配 "key": { 模式（block 开始）
		if idx := strings.Index(trimmed, `":`); idx > 0 && strings.Contains(trimmed[idx:], "{") {
			key := strings.Trim(trimmed[:idx], `" `)
			// 在当前行及后续 4 行中查找注释
			for j := i; j < len(lines) && j < i+5; j++ {
				c := extractInlineComment(lines[j])
				if c != "" {
					comments[key] = c
					break
				}
				// 遇到 } 且没有 {（即纯闭合行）则 block 结束
				if strings.Contains(lines[j], "}") && !strings.Contains(lines[j], "{") {
					break
				}
			}
		}
	}
	return comments
}

// extractInlineComment 提取一行中 // 后的注释文本。
func extractInlineComment(line string) string {
	idx := strings.Index(line, "//")
	if idx < 0 {
		return ""
	}
	before := line[:idx]
	// 确保 // 在引号外（偶数个引号）
	if strings.Count(before, `"`)%2 != 0 {
		return ""
	}
	return strings.TrimSpace(line[idx+2:])
}

// deriveLabel 从注释文本中提取 ≤5 字的中文简称。
// 策略：按中文标点截断取第一段 → 超出5字取前5字。
func deriveLabel(comment string) string {
	if comment == "" {
		return ""
	}

	// 尝试在分隔符处截断
	for _, sep := range []string{"：", ":", "，", "、", "；", "；", "——"} {
		if idx := strings.Index(comment, sep); idx > 0 {
			first := strings.TrimSpace(comment[:idx])
			runes := []rune(first)
			if len(runes) <= 5 && len(runes) > 0 {
				return first
			}
		}
	}

	// 无分隔符或第一段仍超长：取前5字符
	runes := []rune(comment)
	if len(runes) > 5 {
		return string(runes[:5])
	}
	return comment
}
