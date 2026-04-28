package main

import (
	"os/exec"
	"strings"
	"syscall"
)

// runModelsCmd 执行 opencode models 命令，返回模型 ID 列表。
func runModelsCmd() ([]string, error) {
	cmd := exec.Command("opencode", "models")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	models := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			models = append(models, line)
		}
	}
	return models, nil
}
