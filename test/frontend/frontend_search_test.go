package frontendtest

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSearchNavigationUsesTemporaryExpansionWithoutPersistingExpandedParts(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../frontend/dist/chat.js")
	if err != nil {
		t.Fatalf("读取聊天脚本失败: %v", err)
	}
	source := string(sourceBytes)

	for _, required := range []string{
		"let searchTemporaryExpansion = null;",
		"restoreSearchTemporaryExpansion();",
		"temporarilyRevealSearchResult(current);",
		"function temporarilyRevealSearchResult(node) {",
		"function restoreSearchTemporaryExpansion() {",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("缺少搜索临时展开机制: %s", required)
		}
	}

	re := regexp.MustCompile(`(?s)function temporarilyRevealSearchResult\(node\) \{(.*?)\n\}`)
	match := re.FindStringSubmatch(source)
	if len(match) != 2 {
		t.Fatal("未找到 temporarilyRevealSearchResult 函数体")
	}
	if strings.Contains(match[1], "expandedParts[") {
		t.Fatal("搜索临时展开不能写入 expandedParts，避免多次搜索导致卡片永久展开")
	}
}

func TestSessionTooltipIncludesTitleDirectoryAndUpdatedAt(t *testing.T) {
	chatBytes, err := os.ReadFile("../../frontend/dist/chat.js")
	if err != nil {
		t.Fatalf("读取聊天脚本失败: %v", err)
	}
	chatSource := string(chatBytes)

	for _, required := range []string{
		`oc-tree-tooltip`,
		`ses.title`,
		`updatedAt`,
		`sesDir`,
		`oc-tree-tooltip-title`,
		`oc-tree-tooltip-row`,
	} {
		if !strings.Contains(chatSource, required) {
			t.Fatalf("聊天脚本缺少会话气泡元素: %s", required)
		}
	}

	cssBytes, err := os.ReadFile("../../frontend/dist/style.css")
	if err != nil {
		t.Fatalf("读取样式表失败: %v", err)
	}
	cssSource := string(cssBytes)

	for _, required := range []string{
		`.oc-tree-tooltip {`,
		`.oc-tree-tooltip-title {`,
		`.oc-tree-tooltip-row {`,
		`.oc-tree-session:hover .oc-tree-tooltip`,
		`.oc-client`,
	} {
		if !strings.Contains(cssSource, required) {
			t.Fatalf("样式表缺少气泡样式: %s", required)
		}
	}
	if strings.Contains(cssSource, `.oc-client`) {
		if strings.Contains(cssSource, `.oc-client {\n`) {
		}
	}
}

func TestProjectTreeRefreshOnlyOnStructuralSessionEvents(t *testing.T) {
	chatBytes, err := os.ReadFile("../../frontend/dist/chat.js")
	if err != nil {
		t.Fatalf("读取聊天脚本失败: %v", err)
	}
	chatSource := string(chatBytes)

	for _, required := range []string{
		`if (type === 'session.deleted') {`,
		`buildTree();`,
		`const isCurrentSession = sid && sid === currentSessionId;`,
		`if (type === 'session.created' && isCurrentSession) {`,
	} {
		if !strings.Contains(chatSource, required) {
			t.Fatalf("项目树缺少结构性会话事件刷新逻辑: %s", required)
		}
	}

	forbiddenPatterns := map[string]*regexp.Regexp{
		"session.status idle 刷新项目树": regexp.MustCompile(`(?s)if \(type === 'session\.status' && sid\) \{.*?if \(status\?\.type === 'idle'\) \{\s*loadMessages\(\);\s*debounceRefreshTree\(\);`),
		"session.idle 刷新项目树":        regexp.MustCompile(`(?s)if \(type === 'session\.idle' && sid\) \{.*?loadMessages\(\);\s*debounceRefreshTree\(\);`),
		"session.updated 刷新项目树":     regexp.MustCompile(`(?s)if \(type === 'session\.updated'\) \{\s*loadDiff\(\);\s*debounceRefreshTree\(\);`),
	}

	for name, pattern := range forbiddenPatterns {
		if pattern.MatchString(chatSource) {
			t.Fatalf("会话过程事件不应触发项目树刷新: %s", name)
		}
	}

	createdRefreshRe := regexp.MustCompile(`(?s)if \(type === 'session\.created' \|\| type === 'session\.deleted'\) \{\s*buildTree\(\);`)
	if createdRefreshRe.MatchString(chatSource) {
		t.Fatal("session.created 不应再无条件触发项目树刷新")
	}
}

func TestProviderSaveIncludesNpmSelection(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/provider-view.js")
	if err != nil {
		t.Fatalf("读取供应商脚本失败: %v", err)
	}
	source := string(js)

	for _, required := range []string{
		"const npmSelect = card.querySelector('.prov-edit-npm');",
		"const selectedNpm = npmSelect?.value || '';",
		"const rawNpm = npmSelect?.dataset.rawNpm || '';",
		"npm: selectedNpm === PROVIDER_NPM_UNMATCHED ? rawNpm : selectedNpm,",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("供应商脚本缺少接口格式保存线索: %s", required)
		}
	}
}

func TestProviderViewIncludesNpmFormatOptions(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/provider-view.js")
	if err != nil {
		t.Fatalf("读取供应商脚本失败: %v", err)
	}
	source := string(js)

	for _, required := range []string{
		"const PROVIDER_NPM_OPTIONS = [",
		"@ai-sdk/openai",
		"@ai-sdk/openai-compatible",
		"@ai-sdk/anthropic",
		"@ai-sdk/amazon-bedrock",
		"@ai-sdk/google",
		"OpenAI Responses",
		"OpenAI Compatible",
		"Anthropic",
		"Amazon Bedrock",
		"Google (Gemini)",
		"const PROVIDER_NPM_UNMATCHED = '__unmatched__';",
		"未匹配保留",
		"npm: '@ai-sdk/openai-compatible'",
		"npmRaw: '@ai-sdk/openai-compatible'",
		"const matchedOption = PROVIDER_NPM_OPTIONS.find(item => item.value === (p.npm || ''));",
		"const selectedNpmValue = matchedOption ? matchedOption.value : PROVIDER_NPM_UNMATCHED;",
		"<option value=\"${PROVIDER_NPM_UNMATCHED}\" selected>未匹配保留</option>",
		"data-raw-npm=\"${escapeHtml(p.npm || '')}\"",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("供应商脚本缺少接口格式选项: %s", required)
		}
	}
}

func TestOmoSchemeSaveMergesHiddenFields(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/omo-config.js")
	if err != nil {
		t.Fatalf("读取 OMO 脚本失败: %v", err)
	}
	source := string(js)

	for _, required := range []string{
		"let saveBaseConfigJson = {};",
		"function buildMergedConfigForSave() {",
		"const merged = JSON.parse(JSON.stringify(saveBaseConfigJson || {}));",
		"for (const type of modelTypes) {",
		"delete merged[type];",
		"JSON.stringify(buildMergedConfigForSave(), null, 2)",
		"api.SaveFullConfig(",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("OMO 保存脚本缺少保存合并逻辑: %s", required)
		}
	}
}

func TestOmoSchemeSwitchUpdatesSaveBaseConfig(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/omo-config.js")
	if err != nil {
		t.Fatalf("读取 OMO 脚本失败: %v", err)
	}
	source := string(js)

	for _, required := range []string{
		"saveBaseConfigJson = JSON.parse(JSON.stringify(fullConfigJson || {}));",
		"saveBaseConfigJson = JSON.parse(JSON.stringify(data || {}));",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("OMO 方案切换后未更新保存基底: %s", required)
		}
	}
}

func TestOmoSaveButtonUsesHandleSchemeApply(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/main.js")
	if err != nil {
		t.Fatalf("读取 main.js 失败: %v", err)
	}
	source := string(js)

	for _, required := range []string{
		"if (typeof handleSchemeApply === 'function') {",
		"await handleSchemeApply();",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("OMO 保存按钮未统一走 handleSchemeApply: %s", required)
		}
	}

	forbidden := regexp.MustCompile(`(?s)document\.getElementById\('modelActions'\)\.addEventListener\('click', async \(e\) => \{.*?const result = await api\.UpdateModels\(modelEntries\);`)
	if forbidden.MatchString(source) {
		t.Fatal("OMO 保存按钮不应再直接调用 api.UpdateModels(modelEntries)")
	}
}

func TestOmoSaveApplyKeepsButtonFeedback(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/omo-config.js")
	if err != nil {
		t.Fatalf("读取 OMO 脚本失败: %v", err)
	}
	source := string(js)
	lowerSource := strings.ToLower(source)

	for _, required := range []string{
		"showtoast('保存中...', 'info');",
		"btn.disabled = true;",
		"btn.textcontent = '⏳ 保存中...';",
		"btn.disabled = false;",
		"btn.textcontent = '💾 保存';",
		"已保存 ",
		"rendermodelconfig();",
	} {
		if !strings.Contains(lowerSource, strings.ToLower(required)) {
			t.Fatalf("OMO 保存交互缺少线索: %s", required)
		}
	}
}

func TestOmoSchemeSwitchDoesNotSwallowLoadErrors(t *testing.T) {
	js, err := os.ReadFile("../../frontend/dist/omo-config.js")
	if err != nil {
		t.Fatalf("读取 OMO 脚本失败: %v", err)
	}
	source := string(js)

	for _, required := range []string{
		"showToast('方案加载失败: ' + (e.message || e), 'error');",
		"return true;",
		"return false;",
		"await loadSchemeIntoEditor(name);",
		"if (ok) {",
		"showToast('已加载方案: ' + name, 'success');",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("方案切换缺少失败传播或成功保护逻辑: %s", required)
		}
	}
}
