package main

import (
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

var (
	cachedModels   []string
	cachedModelsMu sync.RWMutex
	cachedModelsOk bool
)

// loadModels 后台加载可用模型列表（线程安全，仅执行一次查询）
func loadModels() {
	cachedModelsMu.Lock()
	defer cachedModelsMu.Unlock()
	if cachedModelsOk {
		return
	}
	models, err := fetchModels()
	if err == nil && len(models) > 0 {
		cachedModels = models
		cachedModelsOk = true
	}
}

// refreshModels 强制刷新可用模型列表
func refreshModels() {
	cachedModelsMu.Lock()
	defer cachedModelsMu.Unlock()
	models, err := fetchModels()
	if err == nil && len(models) > 0 {
		cachedModels = models
		cachedModelsOk = true
	}
}

// getAvailableModels 获取已缓存的可用模型列表
func getAvailableModels() ([]string, error) {
	cachedModelsMu.RLock()
	defer cachedModelsMu.RUnlock()
	if !cachedModelsOk {
		// 如果缓存为空，尝试首次加载
		cachedModelsMu.RUnlock()
		loadModels()
		cachedModelsMu.RLock()
	}
	if cachedModelsOk {
		return cachedModels, nil
	}
	// 紧急回退：直接执行命令
	return fetchModels()
}

// fetchModels 执行 opencode models 命令，返回模型 ID 列表。
func fetchModels() ([]string, error) {
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

// runModelsCmd 保留兼容旧方法名
func runModelsCmd() ([]string, error) {
	return fetchModels()
}
