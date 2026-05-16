// Package config 测试辅助桥接——为 test/ 目录下的外部测试导出内部符号。
// 函数名中的 Test 前缀标识其仅用于测试场景。

package config

import (
	"os"

	"oc-manager/model"
)

// TestLoadOpenCodeConfig 暴露 loadOpenCodeConfig 供外部测试使用。
func TestLoadOpenCodeConfig() (*model.OpenCodeConfig, error) {
	return loadOpenCodeConfig()
}

// TestWriteConfigFile 暴露 writeConfigFile 供外部测试使用。
func TestWriteConfigFile(path string, data []byte, perm os.FileMode) error {
	return writeConfigFile(path, data, perm)
}

// TestValidateJSONC 暴露 validateJSONC 供外部测试使用。
func TestValidateJSONC(data []byte) error {
	return validateJSONC(data)
}

// TestParseModelConfigSections 暴露 parseModelConfigSections 供外部测试使用。
func TestParseModelConfigSections(cleaned string) (model.OpenAgentConfig, error) {
	return parseModelConfigSections(cleaned)
}

// TestStripComments 暴露 stripComments 供外部测试使用。
func TestStripComments(text string) string {
	return stripComments(text)
}
