package skill

import (
	"fmt"
	"strings"
)

// skillFrontmatter SKILL.md frontmatter 解析结果。
type skillFrontmatter struct {
	Name        string
	Description string
}

func parseFrontmatter(data []byte) (skillFrontmatter, error) {
	content := string(data)
	if len(content) < 4 || content[:3] != "---" {
		return skillFrontmatter{}, fmt.Errorf("无 frontmatter")
	}
	rest := content[3:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		end = strings.Index(rest, "---")
	}
	if end == -1 {
		return skillFrontmatter{}, fmt.Errorf("frontmatter 未闭合")
	}
	yamlText := rest[:end]

	var fm skillFrontmatter
	lines := strings.Split(yamlText, "\n")
	var currentKey string
	var currentValue []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) && currentKey != "" {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				currentValue = append(currentValue, trimmed)
			}
			continue
		}
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

		if currentKey != "" {
			switch currentKey {
			case "name":
				fm.Name = strings.Join(currentValue, " ")
			case "description":
				fm.Description = strings.Join(currentValue, " ")
			}
		}
		currentKey = key
		currentValue = []string{val}
	}

	if currentKey != "" {
		switch currentKey {
		case "name":
			fm.Name = strings.Join(currentValue, " ")
		case "description":
			fm.Description = strings.Join(currentValue, " ")
		}
	}

	return fm, nil
}
