# OC Manager 功能增强 — 实施计划

> **用于 agentic workers**: 请使用 superpowers:subagent-driven-development（推荐）或 superpowers:executing-plans 逐任务实施此计划。步骤使用 `- [ ]` 复选框语法追踪。

**目标**: 在现有 OC Manager 基础上增量实现四项改进：工作区交互体验优化、模型配置改名 OMO、技能管理重构、配置文件扩展名兼容。

**架构**: 纯增量改动，不碰工作区核心消息渲染逻辑。前端改动集中在 chat.js（+~80行）、skill-manager.js（重写）、index.html（+~30行 DOM），后端改动集中在 skill/skill.go（重写）和 config/（+~18行 fallback 逻辑）。

**技术栈**: Go 1.25 + Wails v2.12 + 原生 HTML/CSS/JS + marked.js

---

## 文件结构映射

| 文件 | 职责 | 改动类型 |
|------|------|---------|
| `config/model_config.go` | 模型配置读写 + 路径解析 | 修改 ~10行 |
| `config/provider_config.go` | 供应商配置读写 + 路径解析 | 修改 ~8行 |
| `skill/skill.go` | 技能扫描/链接管理 | 重写 |
| `model/types.go` | 共享数据类型 | 修改 ~20行 |
| `app.go` | Wails 绑定方法 | 修改 ~15行 |
| `frontend/dist/index.html` | 页面结构 | 修改 ~30行 |
| `frontend/dist/chat.js` | 工作区核心逻辑 | 修改 ~80行 |
| `frontend/dist/navigation.js` | 侧边栏导航 | 修改 ~5行 |
| `frontend/dist/skill-manager.js` | 技能管理视图 | 重写 |
| `frontend/dist/omo-config.js` | OMO 配置视图（改自 model-config.js） | 改名+文案修改 |
| `frontend/dist/style.css` | 样式表 | 修改 ~60行 |

---

### Task 1: 配置文件扩展名兼容

**文件**: 
- 修改: `config/model_config.go:20-24`
- 修改: `config/provider_config.go:14-18`

- [ ] **Step 1: 在 config 包中添加 resolvePath 函数**

在 `config/model_config.go` 文件末尾（`package config` 内部）添加：

```go
// resolvePath 优先返回 .jsonc 路径，若不存在则回退到 .json。
// 两者都不存在时返回原始 .jsonc 路径，由调用方处理错误。
func resolvePath(jsoncPath string) string {
	if _, err := os.Stat(jsoncPath); err == nil {
		return jsoncPath
	}
	jsonPath := strings.TrimSuffix(jsoncPath, ".jsonc") + ".json"
	if _, err := os.Stat(jsonPath); err == nil {
		return jsonPath
	}
	return jsoncPath
}
```

需要添加的 import：`"strings"`（检查 model_config.go 是否已导入，当前第 1-17 行无 strings，需要添加）。

- [ ] **Step 2: 修改 ConfigPath() 使用 resolvePath**

修改 `config/model_config.go` 第 20-24 行：

```go
// ConfigPath 返回 oh-my-openagent 配置文件的完整路径。
// 优先 .jsonc，不存在时回退到 .json。
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return resolvePath(filepath.Join(home, ".config", "opencode", "oh-my-openagent.jsonc"))
}
```

- [ ] **Step 3: 修改 OpenCodeConfigPath() 使用 resolvePath**

修改 `config/provider_config.go` 第 14-18 行：

```go
// OpenCodeConfigPath 返回 opencode 配置文件的完整路径。
// 优先 .jsonc，不存在时回退到 .json。
func OpenCodeConfigPath() string {
	home, _ := os.UserHomeDir()
	return resolvePath(filepath.Join(home, ".config", "opencode", "opencode.jsonc"))
}
```

- [ ] **Step 4: 添加 strings import**

确认 `config/model_config.go` imports 中已包含 `"strings"`。阅读文件头部确认，若无则添加。

- [ ] **Step 5: 编译验证**

```bash
go build ./...
```

预期: 编译通过。

- [ ] **Step 6: 运行测试**

```bash
go test ./...
```

预期: 所有测试通过。

- [ ] **Step 7: 提交**

```cmd
git add config/model_config.go config/provider_config.go
git commit -m "fix: 配置文件兼容 .jsonc/.json 双扩展名"
```

---

### Task 2: 模型配置 → OMO 配置改名

**文件**:
- 重命名: `frontend/dist/model-config.js` → `frontend/dist/omo-config.js`
- 修改: `frontend/dist/index.html:31-33,148-165,270-272`
- 修改: `frontend/dist/navigation.js:2-26`

- [ ] **Step 1: 重命名 model-config.js 为 omo-config.js**

```cmd
move frontend\dist\model-config.js frontend\dist\omo-config.js
```

- [ ] **Step 2: 修改 omo-config.js 内部文案**

打开 `frontend/dist/omo-config.js`，将文件内的所有 `模型配置` 替换为 `OMO 配置`（包括页面标题、loading 文字等）。

- [ ] **Step 3: 修改 index.html 中的导航项文案**

修改第 33-35 行的侧边栏导航项：

```html
<div class="nav-item" data-view="view-omo">
    <span class="nav-icon">⚙️</span>
    <span class="nav-label">OMO 配置</span>
</div>
```

替换原来的 `data-view="view-models"` 和 `模型配置`。

- [ ] **Step 4: 修改 index.html 中的视图面板 ID**

修改第 149 行，把面板 ID 从 `view-models` 改为 `view-omo`：

```html
<div class="view-panel" id="view-omo">
```

- [ ] **Step 5: 修改 index.html 中的 script 引用**

修改第 272 行（靠近底部 script 加载区）：

```html
<script src="/omo-config.js"></script>
```

替换原来的 `<script src="/model-config.js"></script>`。

- [ ] **Step 6: 修改 navigation.js 中的路由**

修改 `frontend/dist/navigation.js` 第 16 行附近：

```js
} else if (viewId === 'view-omo') {
    loadModelConfig();
```

替换原来的 `'view-models'`。

- [ ] **Step 7: 修改 main.js 中的视图引用**

搜索 `main.js` 中所有 `view-models` 引用并改为 `view-omo`。

- [ ] **Step 8: 编译验证**

```bash
go build ./...
```

预期: 编译通过（前端文件由 Wails embed，需确保新文件名被包含）。

- [ ] **Step 9: 提交**

```cmd
git add -A
git commit -m "refactor: 模型配置改名为 OMO 配置"
```

---

### Task 3: 技能管理后端重构

**文件**:
- 重写: `skill/skill.go`
- 修改: `model/types.go`
- 修改: `app.go`

- [ ] **Step 1: 更新 model/types.go — 清理旧类型并新增 SkillContent**

先读取 `model/types.go`，然后：

**删除**以下类型及其全部字段：
- `TargetInfo` (第 15-19 行)
- `Stats.TargetStats` 字段（第 24 行，保留 `TotalSkills`）
- `BatchResult` (第 37-42 行)
- `ToggleResult.Target` 字段（第 30 行）

**新增** `SkillContent` 类型：

```go
// SkillContent 技能文件内容。
type SkillContent struct {
	Path    string `json:"path"`    // 技能目录路径
	Content string `json:"content"` // SKILL.md 文件内容
}
```

**更新 `SkillInfo`**：
```go
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`        // 技能目录完整路径
	Linked      bool   `json:"linked"`      // 是否已启用（符号链接存在）
	Source      string `json:"source"`      // "global" 或 "project"
}
```

**更新 `Stats`**：
```go
type Stats struct {
	GlobalSkills  int `json:"globalSkills"`
	ProjectSkills int `json:"projectSkills"`
}
```

- [ ] **Step 2: 重写 skill/skill.go**

完整重写文件（约 200 行），核心逻辑：

```go
package skill

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"oc-manager/model"
)

// skillFrontmatter 解析 SKILL.md 中的 YAML frontmatter（手动解析，无外部依赖）。
type skillFrontmatter struct {
	Name        string
	Description string
}

// NewManager 创建新的 Manager 实例。
func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	return &Manager{
		globalDir: filepath.Join(homeDir, ".config", "opencode", "skills"),
	}
}

// SetProjectDir 设置项目技能目录。
func (m *Manager) SetProjectDir(projectRoot string) {
	m.projectDir = filepath.Join(projectRoot, ".opencode", "skills")
}

// SourceDir 返回全局技能目录路径。
func (m *Manager) SourceDir() string {
	return m.globalDir
}

// skillFrontmatter 解析 SKILL.md 中的 YAML frontmatter。
type skillFrontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// GetAllSkills 扫描全局+项目技能目录，返回所有技能信息。
func (m *Manager) GetAllSkills() []model.SkillInfo {
	var skills []model.SkillInfo

	// 扫描全局技能
	globalSkills := m.scanDir(m.globalDir, "global")
	skills = append(skills, globalSkills...)

	// 扫描项目技能（如果设置了项目目录且存在）
	if m.projectDir != "" {
		if info, err := os.Stat(m.projectDir); err == nil && info.IsDir() {
			projectSkills := m.scanDir(m.projectDir, "project")
			skills = append(skills, projectSkills...)
		}
	}

	return skills
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

		// 检测链接状态：检查 opencode skills 目录下是否存在该符号链接
		// opencode skills 目录 = globalDir（即 ~/.config/opencode/skills/）
		linked := false
		linkPath := filepath.Join(m.globalDir, skillName)
		if info, err := os.Lstat(linkPath); err == nil {
			linked = (info.Mode()&os.ModeSymlink != 0) || info.IsDir() // Windows junction 可能表现为目录
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

// ReadSkillContent 读取指定技能的 SKILL.md 完整内容。
func (m *Manager) ReadSkillContent(skillPath string) (string, error) {
	mdPath := filepath.Join(skillPath, "SKILL.md")
	data, err := os.ReadFile(mdPath)
	if err != nil {
		return "", fmt.Errorf("读取 SKILL.md 失败: %w", err)
	}
	return string(data), nil
}

// SaveSkillContent 保存 SKILL.md 内容。
func (m *Manager) SaveSkillContent(skillPath, content string) error {
	mdPath := filepath.Join(skillPath, "SKILL.md")
	return os.WriteFile(mdPath, []byte(content), 0644)
}

// ToggleSkill 切换技能链接状态。
// enable=true 时创建符号链接到 opencode skills 目录；enable=false 时删除链接。
func (m *Manager) ToggleSkill(skillPath, skillName string, enable bool) (bool, error) {
	linkPath := filepath.Join(m.globalDir, skillName)

	if enable {
		// 创建符号链接
		if runtime.GOOS == "windows" {
			// Windows: 使用 mklink /J（目录联接，无需管理员权限）
			cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, skillPath)
			if err := cmd.Run(); err != nil {
				// 降级到 os.Symlink（需要管理员权限或开发者模式）
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
		// 删除符号链接
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			return false, fmt.Errorf("删除符号链接失败: %w", err)
		}
		return false, nil
	}
}

// parseFrontmatter 解析 Markdown 文件的 YAML frontmatter（--- ... ---）。
// 使用简单字符串解析，不引入外部 YAML 库。
func parseFrontmatter(data []byte) (skillFrontmatter, error) {
	content := string(data)
	if len(content) < 4 || content[:3] != "---" {
		return skillFrontmatter{}, fmt.Errorf("无 frontmatter")
	}
	// 在第二个 "---" 处截断
	rest := content[3:]
	end := strings.Index(rest, "---")
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
		// 去除引号
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
```

注意：`strings` 包需要在 `skill.go` import 中确认存在。

- [ ] **Step 3: 更新 app.go — 技能管理绑定方法**

修改 `app.go` 中的技能管理部分（第 49-131 行）：

**删除**以下方法：
- `GetTargets()` — 不再有多平台目标
- `ToggleAllSkills()` — 不再有跨平台批量操作

**新增**以下方法：

```go
// SetProjectSkillsDir 设置项目技能目录。
func (a *App) SetProjectSkillsDir(projectRoot string) {
	a.sm.SetProjectDir(projectRoot)
}

// GetProjectSkills 获取项目级技能列表。
func (a *App) GetProjectSkills(projectRoot string) []model.SkillInfo {
	a.sm.SetProjectDir(projectRoot)
	return a.sm.GetAllSkills()
}

// ReadSkillContent 读取技能 SKILL.md 内容。
func (a *App) ReadSkillContent(skillPath string) (string, error) {
	return a.sm.ReadSkillContent(skillPath)
}

// SaveSkillContent 保存技能 SKILL.md 内容。
func (a *App) SaveSkillContent(skillPath, content string) error {
	return a.sm.SaveSkillContent(skillPath, content)
}
```

**修改** `ToggleSkill` 签名（移除 `target` 参数）：

```go
func (a *App) ToggleSkill(skillPath, skillName string, enable bool) model.ToggleResult {
	newState, err := a.sm.ToggleSkill(skillPath, skillName, enable)
	// ... 其余不变
}
```

- [ ] **Step 4: 更新 Stats 计算逻辑**

修改 `app.go` 中的 `GetStats()` 方法（第 77-94 行）：

```go
func (a *App) GetStats() model.Stats {
	skills := a.sm.GetAllSkills()
	s := model.Stats{}
	for _, sk := range skills {
		if sk.Source == "global" {
			s.GlobalSkills++
		} else {
			s.ProjectSkills++
		}
	}
	return s
}
```

- [ ] **Step 5: 编译验证**

```bash
go build ./...
```

预期: 编译通过，修复所有编译错误。

- [ ] **Step 6: 运行测试（可能会失败——技能管理无现有测试）**

```bash
go test ./...
```

- [ ] **Step 7: 提交**

```cmd
git add skill/skill.go model/types.go app.go
git commit -m "refactor: 技能管理重构为 opencode 目录扫描 + SKILL.md 读写"
```

---

### Task 4: 技能管理前端重构

**文件**:
- 重写: `frontend/dist/skill-manager.js`
- 修改: `frontend/dist/style.css`（新增 Modal 样式）

- [ ] **Step 1: 重写 skill-manager.js**

完整重写 `frontend/dist/skill-manager.js`（约 200 行）：

```js
// OpenCode 管理中心 - 技能管理视图（重构版）
let skills = [];
let skillsLoaded = false;

async function loadSkillsData() {
    if (skillsLoaded) return;
    skillsLoaded = true;

    try {
        // 获取全局技能
        skills = await api.GetSkills();
        // 获取项目技能（可选：由用户选择项目目录）
        const stats = await api.GetStats();
        renderStats(stats);
        renderSkillList();
        registerSkillEvents();
    } catch (err) {
        skillsLoaded = false;
        showToast('加载技能数据失败: ' + (err.message || err), 'error');
    }
}

function renderStats(stats) {
    document.getElementById('statGlobal').textContent = stats.globalSkills || 0;
    document.getElementById('statProject').textContent = stats.projectSkills || 0;
}

function renderSkillList(filter = '') {
    const list = document.getElementById('skillList');
    if (!skills.length) {
        list.innerHTML = '<div class="oc-empty">暂无技能</div>';
        return;
    }

    const filtered = filter
        ? skills.filter(s => s.name.toLowerCase().includes(filter.toLowerCase()))
        : skills;

    list.innerHTML = filtered.map(s => {
        const sourceLabel = s.source === 'project' ? '项目' : '全局';
        const sourceClass = s.source === 'project' ? 'skill-source-project' : 'skill-source-global';
        return `
        <div class="skill-card" data-skill="${escapeHtml(s.name)}" data-path="${escapeHtml(s.path)}">
            <div class="skill-info">
                <div class="skill-name-row">
                    <span class="skill-name" role="button" tabindex="0"
                          onclick="previewSkill('${escapeHtml(s.path)}')">${escapeHtml(s.name)}</span>
                    <span class="skill-tag ${sourceClass}">${sourceLabel}</span>
                </div>
                <div class="skill-desc">${escapeHtml(s.description || '无描述')}</div>
                <div class="skill-path">${escapeHtml(s.path)}</div>
            </div>
            <div class="skill-actions">
                <button class="btn btn-sm" onclick="editSkill('${escapeHtml(s.path)}')">编辑</button>
                <label class="skill-toggle">
                    <input type="checkbox" ${s.linked ? 'checked' : ''}
                           onchange="toggleSkill('${escapeHtml(s.path)}', '${escapeHtml(s.name)}', this.checked)">
                    <span class="skill-toggle-slider"></span>
                </label>
            </div>
        </div>`;
    }).join('');
}

function registerSkillEvents() {
    // 搜索框
    const searchInput = document.getElementById('skillSearch');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            renderSkillList(e.target.value);
        });
    }
}

// ========== 预览 Modal ==========

async function previewSkill(skillPath) {
    try {
        const result = await api.ReadSkillContent(skillPath);
        const html = marked.parse(result);
        showSkillModal(html, skillPath, false);
    } catch (err) {
        showToast('读取技能失败: ' + (err.message || err), 'error');
    }
}

async function editSkill(skillPath) {
    try {
        const result = await api.ReadSkillContent(skillPath);
        showSkillModal(result, skillPath, true);
    } catch (err) {
        showToast('读取技能失败: ' + (err.message || err), 'error');
    }
}

function showSkillModal(content, skillPath, isEdit) {
    const modal = document.getElementById('skillModal');
    const body = document.getElementById('skillModalBody');
    const title = document.getElementById('skillModalTitle');
    const editBtn = document.getElementById('skillModalEdit');
    const saveBtn = document.getElementById('skillModalSave');
    const cancelBtn = document.getElementById('skillModalCancel');

    title.textContent = skillPath.split(/[/\\]/).pop();
    modal.dataset.skillPath = skillPath;

    if (isEdit) {
        body.innerHTML = `<textarea id="skillEditArea" class="skill-edit-area">${escapeHtml(content)}</textarea>`;
        editBtn.style.display = 'none';
        saveBtn.style.display = '';
        cancelBtn.style.display = '';
    } else {
        body.innerHTML = content; // marked 渲染后的 HTML
        editBtn.style.display = '';
        saveBtn.style.display = 'none';
        cancelBtn.style.display = 'none';
    }

    modal.style.display = 'flex';
}

function closeSkillModal() {
    document.getElementById('skillModal').style.display = 'none';
}

async function saveSkillEdit() {
    const modal = document.getElementById('skillModal');
    const skillPath = modal.dataset.skillPath;
    const content = document.getElementById('skillEditArea').value;

    try {
        await api.SaveSkillContent(skillPath, content);
        showToast('保存成功', 'success');
        // 切回预览
        const html = marked.parse(content);
        showSkillModal(html, skillPath, false);
    } catch (err) {
        showToast('保存失败: ' + (err.message || err), 'error');
    }
}

// 关闭 Modal 事件
document.getElementById('skillModalClose')?.addEventListener('click', closeSkillModal);
document.getElementById('skillModal')?.addEventListener('click', (e) => {
    if (e.target.id === 'skillModal') closeSkillModal();
});
document.getElementById('skillModalEdit')?.addEventListener('click', () => {
    editSkill(document.getElementById('skillModal').dataset.skillPath);
});
document.getElementById('skillModalSave')?.addEventListener('click', saveSkillEdit);
document.getElementById('skillModalCancel')?.addEventListener('click', () => {
    const modal = document.getElementById('skillModal');
    previewSkill(modal.dataset.skillPath);
});

// ========== Toggle 开关 ==========

async function toggleSkill(skillPath, skillName, enable) {
    try {
        const result = await api.ToggleSkill(skillPath, skillName, enable);
        if (result.success) {
            showToast(`已${enable ? '启用' : '禁用'} ${skillName}`, 'success');
            skillsLoaded = false;
            loadSkillsData();
        } else {
            showToast('操作失败: ' + (result.error || '未知错误'), 'error');
        }
    } catch (err) {
        showToast('操作失败: ' + (err.message || err), 'error');
    }
}
```

- [ ] **Step 2: 更新 index.html 中的技能管理 HTML**

修改 `index.html` 中技能管理面板（第 168-195 行附近），改为：

```html
<!-- ===== View 3: 技能管理 ===== -->
<div class="view-panel" id="view-skills">
    <!-- 统计栏 -->
    <section class="stats-bar">
        <div class="stat-item">
            <span class="stat-value" id="statGlobal">0</span>
            <span class="stat-label">全局技能</span>
        </div>
        <div class="stat-item">
            <span class="stat-value" id="statProject">0</span>
            <span class="stat-label">项目技能</span>
        </div>
        <div class="stats-spacer"></div>
        <button class="btn btn-sm btn-refresh" id="btnRefresh" title="刷新技能列表">🔄 刷新</button>
    </section>

    <!-- 源路径 -->
    <section class="path-bar">
        <span class="path-label">全局目录：</span>
        <span class="path-value" id="sourcePath">加载中...</span>
        <button class="btn btn-sm btn-open" id="btnOpenDir" title="在资源管理器中打开">📂 打开</button>
    </section>

    <!-- 搜索栏 -->
    <section class="path-bar">
        <span class="path-label">搜索：</span>
        <input type="text" id="skillSearch" placeholder="输入技能名称过滤..." style="flex:1;padding:4px 8px;" />
    </section>

    <!-- 技能列表 -->
    <main class="skill-list" id="skillList">
        <div class="loading"><div class="spinner"></div><p>正在加载技能列表...</p></div>
    </main>
</div>
```

同时移除批量操作栏（`batch-bar`）相关 HTML。

- [ ] **Step 3: 添加技能 Modal HTML（在 index.html 底部，Toast 下方）**

```html
<!-- 技能预览/编辑 Modal -->
<div class="modal-overlay" id="skillModal" style="display:none">
    <div class="modal skill-modal">
        <div class="modal-header">
            <h3 id="skillModalTitle">技能名称</h3>
            <div class="modal-header-actions">
                <button class="btn btn-sm" id="skillModalEdit">编辑</button>
                <button class="btn btn-sm btn-primary" id="skillModalSave" style="display:none">保存</button>
                <button class="btn btn-sm btn-cancel" id="skillModalCancel" style="display:none">取消</button>
                <button class="btn btn-sm" id="skillModalClose">×</button>
            </div>
        </div>
        <div class="modal-body" id="skillModalBody"></div>
    </div>
</div>
```

- [ ] **Step 4: 更新 style.css — Modal + 技能卡片样式**

追加以下 CSS：

```css
/* ======== 技能 Modal ======== */
.skill-modal { width: 720px; max-height: 80vh; display: flex; flex-direction: column; }
.skill-modal .modal-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; border-bottom: 1px solid var(--border); }
.skill-modal .modal-header-actions { display: flex; gap: 8px; align-items: center; }
.skill-modal .modal-body { flex: 1; overflow-y: auto; padding: 16px; }
.skill-modal .modal-body h1 { font-size: 1.3em; margin: 8px 0 12px; }
.skill-modal .modal-body h2 { font-size: 1.1em; margin: 8px 0 10px; }
.skill-modal .modal-body pre { background: var(--bg-code); padding: 12px; border-radius: 4px; overflow-x: auto; }
.skill-edit-area { width: 100%; min-height: 400px; font-family: 'Cascadia Code', 'Fira Code', monospace; font-size: 13px; line-height: 1.5; padding: 12px; border: 1px solid var(--border); border-radius: 4px; background: var(--bg-input); color: var(--text); resize: vertical; }

/* ======== 技能卡片 ======== */
.skill-name { color: var(--accent); cursor: pointer; text-decoration: underline; text-underline-offset: 2px; }
.skill-name:hover { color: var(--accent-hover); }
.skill-tag { font-size: 11px; padding: 1px 6px; border-radius: 3px; }
.skill-source-global { background: var(--bg-tag); color: var(--text-dim); }
.skill-source-project { background: color-mix(in srgb, var(--accent) 15%, transparent); color: var(--accent); }
.skill-path { font-size: 11px; color: var(--text-dim); font-family: monospace; margin-top: 4px; }
.skill-toggle { position: relative; display: inline-block; width: 36px; height: 20px; }
.skill-toggle input { opacity: 0; width: 0; height: 0; }
.skill-toggle-slider { position: absolute; inset: 0; background: var(--border); border-radius: 20px; transition: 0.2s; cursor: pointer; }
.skill-toggle-slider::before { content: ''; position: absolute; height: 14px; width: 14px; left: 3px; bottom: 3px; background: #fff; border-radius: 50%; transition: 0.2s; }
.skill-toggle input:checked + .skill-toggle-slider { background: var(--accent); }
.skill-toggle input:checked + .skill-toggle-slider::before { transform: translateX(16px); }
```

- [ ] **Step 5: 编译验证**

```bash
go build ./...
```

- [ ] **Step 6: 提交**

```cmd
git add frontend/dist/skill-manager.js frontend/dist/index.html frontend/dist/style.css
git commit -m "feat: 技能管理重构 — opencode 目录扫描 + SKILL.md 预览编辑"
```

---

### Task 5: 工作区交互体验 — Tab 标签

**文件**:
- 修改: `frontend/dist/index.html:60-73`
- 修改: `frontend/dist/chat.js`
- 修改: `frontend/dist/style.css`

- [ ] **Step 1: 在 index.html 中插入 Tab 栏 HTML**

在 `index.html` 第 74 行 `<main class="oc-chat">` 上方，插入：

```html
<div class="oc-tabs" id="ocTabs"></div>
```

也就是在 `</aside>` 和 `<main class="oc-chat">` 之间。

- [ ] **Step 2: 在 chat.js 中添加 Tab 管理逻辑**

在 `chat.js` 顶部全局变量区（约第 35 行后）添加：

```js
// ========== Tab 管理 ==========
let openTabs = [];  // [{ id, title, active, domEl }]
const MAX_TABS = 8;

function getTabIndex(sessionId) {
    return openTabs.findIndex(t => t.id === sessionId);
}

function openTab(sessionId, title) {
    const existing = getTabIndex(sessionId);
    if (existing >= 0) {
        activateTab(existing);
        return;
    }
    // 超出上限时关闭最旧未激活的
    if (openTabs.length >= MAX_TABS) {
        const oldestInactive = openTabs.findIndex(t => !t.active);
        if (oldestInactive >= 0) {
            closeTab(oldestInactive);
        } else {
            closeTab(0);
        }
    }
    // 创建消息容器（如果尚不存在）
    let msgEl = document.getElementById('oc-msg-' + sessionId);
    if (!msgEl) {
        msgEl = document.createElement('div');
        msgEl.id = 'oc-msg-' + sessionId;
        msgEl.className = 'oc-messages';
        msgEl.style.display = 'none';
        document.getElementById('ocMessages').parentElement.appendChild(msgEl);
    }
    openTabs.push({ id: sessionId, title: title || '新会话', active: false, domEl: msgEl });
    activateTab(openTabs.length - 1);
}

function activateTab(index) {
    openTabs.forEach((t, i) => {
        t.active = (i === index);
        if (t.domEl) t.domEl.style.display = t.active ? '' : 'none';
    });
    // 同步 currentSessionId
    if (openTabs[index]) {
        currentSessionId = openTabs[index].id;
        updateChatTitle(openTabs[index].title);
    }
    renderTabs();
}

function closeTab(index) {
    if (index < 0 || index >= openTabs.length) return;
    openTabs.splice(index, 1);
    // 清理空的 DOM
    if (openTabs.length === 0) {
        currentSessionId = '';
        updateChatTitle('未选择会话');
        document.getElementById('ocMessages').style.display = '';
    } else {
        const nextIndex = Math.min(index, openTabs.length - 1);
        activateTab(nextIndex);
    }
    renderTabs();
}

function renderTabs() {
    const container = document.getElementById('ocTabs');
    if (!container) return;
    container.innerHTML = openTabs.map((t, i) => {
        const activeClass = t.active ? ' active' : '';
        const title = (t.title || '新会话').substring(0, 18);
        return `<div class="oc-tab${activeClass}" data-index="${i}">
            <span class="oc-tab-title" onclick="activateTab(${i})">${escapeHtml(title)}</span>
            <span class="oc-tab-close" onclick="event.stopPropagation(); closeTab(${i})">×</span>
        </div>`;
    }).join('');
}

function updateChatTitle(title) {
    const el = document.getElementById('ocChatTitle');
    if (el) el.textContent = title || '未选择会话';
}
```

- [ ] **Step 3: 修改 selectSession() 调用 openTab()**

在 `chat.js` 中找到 `selectSession()` 函数（约第 x 行），在设置 `currentSessionId` 之前，添加：

```js
openTab(sessionId, sessionName);
```

同时移除原来的 `currentSessionId = sessionId` 赋值（由 `activateTab` 内部处理）。

- [ ] **Step 4: 在 style.css 中添加 Tab 样式**

```css
/* ======== 会话 Tab 标签 ======== */
.oc-tabs { display: flex; align-items: center; gap: 2px; background: var(--bg-panel); border-bottom: 1px solid var(--border); padding: 0 8px; overflow-x: auto; min-height: 32px; }
.oc-tab { display: flex; align-items: center; gap: 4px; padding: 4px 8px 4px 10px; border-radius: 4px 4px 0 0; cursor: default; font-size: 12px; color: var(--text-dim); border: 1px solid transparent; border-bottom: none; white-space: nowrap; }
.oc-tab:hover { background: var(--bg-hover); }
.oc-tab.active { color: var(--text); background: var(--bg); border-color: var(--border); }
.oc-tab-title { cursor: pointer; max-width: 140px; overflow: hidden; text-overflow: ellipsis; }
.oc-tab-close { cursor: pointer; font-size: 14px; line-height: 1; padding: 0 2px; border-radius: 2px; }
.oc-tab-close:hover { background: var(--bg-hover); color: var(--danger); }
```

- [ ] **Step 5: 编译验证**

```bash
go build ./...
```

- [ ] **Step 6: 提交**

```cmd
git add frontend/dist/chat.js frontend/dist/index.html frontend/dist/style.css
git commit -m "feat: 工作区增加会话 Tab 标签管理"
```

---

### Task 6: 工作区交互体验 — 面板拖拽调整

**文件**:
- 修改: `frontend/dist/index.html:50-145`
- 修改: `frontend/dist/chat.js`
- 修改: `frontend/dist/style.css`

- [ ] **Step 1: 在 index.html 中插入拖拽手柄**

在项目树面板的 `</aside>` 之前添加拖拽手柄，在右侧面板的 `<aside class="oc-sidepanel">` 之前添加手柄：

```html
<div class="oc-resize-handle" id="resizeTree" data-target="ocSessions" data-min="160"></div>
```

```html
<div class="oc-resize-handle" id="resizeSide" data-target="ocSidepanel" data-min="200" data-side="left"></div>
```

- [ ] **Step 2: 在 chat.js 中添加拖拽逻辑**

在 `chat.js` 底部添加：

```js
// ========== 面板拖拽 ==========
function initResizeHandles() {
    document.querySelectorAll('.oc-resize-handle').forEach(handle => {
        let startX, startWidth;
        const targetId = handle.dataset.target;
        const minWidth = parseInt(handle.dataset.min) || 160;
        const target = document.getElementById(targetId);
        if (!target) return;

        handle.addEventListener('mousedown', (e) => {
            e.preventDefault();
            startX = e.clientX;
            startWidth = target.offsetWidth;
            handle.classList.add('active');
            document.body.style.cursor = 'col-resize';
            document.body.style.userSelect = 'none';

            const onMove = (e) => {
                const dx = handle.dataset.side === 'left' ? (startX - e.clientX) : (e.clientX - startX);
                const newWidth = Math.max(minWidth, startWidth + dx);
                target.style.width = newWidth + 'px';
                target.style.flex = 'none';
            };
            const onUp = () => {
                handle.classList.remove('active');
                document.body.style.cursor = '';
                document.body.style.userSelect = '';
                // 持久化宽度
                const widths = JSON.parse(localStorage.getItem('oc-panel-widths') || '{}');
                widths[targetId] = target.style.width;
                localStorage.setItem('oc-panel-widths', JSON.stringify(widths));
                document.removeEventListener('mousemove', onMove);
                document.removeEventListener('mouseup', onUp);
            };
            document.addEventListener('mousemove', onMove);
            document.addEventListener('mouseup', onUp);
        });
    });

    // 恢复持久化宽度
    const savedWidths = JSON.parse(localStorage.getItem('oc-panel-widths') || '{}');
    Object.keys(savedWidths).forEach(id => {
        const el = document.getElementById(id);
        if (el) { el.style.width = savedWidths[id]; el.style.flex = 'none'; }
    });
}
```

- [ ] **Step 3: 在 startup 中调用 initResizeHandles()**

在 `chat.js` 的 `checkWebStatus()` 或 `domReady` 回调中，调用 `initResizeHandles()`。

- [ ] **Step 4: 在 style.css 中添加拖拽手柄样式**

```css
/* ======== 拖拽手柄 ======== */
.oc-resize-handle { width: 4px; cursor: col-resize; background: transparent; transition: background 0.15s; flex-shrink: 0; }
.oc-resize-handle:hover, .oc-resize-handle.active { background: var(--accent); }
```

- [ ] **Step 5: 编译验证**

```bash
go build ./...
```

- [ ] **Step 6: 提交**

```cmd
git add frontend/dist/index.html frontend/dist/chat.js frontend/dist/style.css
git commit -m "feat: 左右面板可拖拽调整宽度"
```

---

### Task 7: 工作区交互体验 — 消息搜索

**文件**:
- 修改: `frontend/dist/index.html:75`
- 修改: `frontend/dist/chat.js`
- 修改: `frontend/dist/style.css`

- [ ] **Step 1: 在 index.html 中插入搜索栏**

在消息区顶部，Tab 栏下方、消息列表上方，插入：

```html
<div class="oc-search-bar" id="ocSearchBar" style="display:none">
    <input type="text" id="ocSearchInput" placeholder="搜索当前会话消息..." />
    <span class="oc-search-count" id="ocSearchCount"></span>
    <button class="btn btn-sm" id="ocSearchPrev" disabled>▲</button>
    <button class="btn btn-sm" id="ocSearchNext" disabled>▼</button>
    <button class="btn btn-sm" id="ocSearchClose">×</button>
</div>
```

- [ ] **Step 2: 在 chat.js 中添加搜索逻辑**

在 `chat.js` 底部添加：

```js
// ========== 消息搜索 ==========
let searchResults = [];
let searchIndex = -1;

function initSearch() {
    const bar = document.getElementById('ocSearchBar');
    const input = document.getElementById('ocSearchInput');
    const countEl = document.getElementById('ocSearchCount');
    const prevBtn = document.getElementById('ocSearchPrev');
    const nextBtn = document.getElementById('ocSearchNext');

    // Ctrl+F 打开
    document.addEventListener('keydown', (e) => {
        if (e.ctrlKey && e.key === 'f') {
            e.preventDefault();
            bar.style.display = 'flex';
            input.focus();
        }
        if (e.key === 'Escape') {
            closeSearch();
        }
    });

    // 输入即搜
    let searchTimer;
    input.addEventListener('input', () => {
        clearTimeout(searchTimer);
        searchTimer = setTimeout(() => doSearch(input.value), 200);
    });

    document.getElementById('ocSearchClose').addEventListener('click', closeSearch);
    prevBtn.addEventListener('click', () => navigateSearch(-1));
    nextBtn.addEventListener('click', () => navigateSearch(1));
}

function doSearch(query) {
    clearHighlights();
    searchResults = [];
    searchIndex = -1;
    document.getElementById('ocSearchCount').textContent = '';
    document.getElementById('ocSearchPrev').disabled = true;
    document.getElementById('ocSearchNext').disabled = true;

    if (!query || query.length < 2) return;

    const msgContainer = document.getElementById('ocMessages');
    if (!msgContainer) return;

    const walker = document.createTreeWalker(msgContainer, NodeFilter.SHOW_TEXT, null, false);
    const nodes = [];
    while (walker.nextNode()) {
        nodes.push(walker.currentNode);
    }

    const lowerQuery = query.toLowerCase();
    nodes.forEach(node => {
        const text = node.textContent.toLowerCase();
        if (text.includes(lowerQuery)) {
            const parent = node.parentElement;
            if (parent) {
                const range = document.createRange();
                const idx = text.indexOf(lowerQuery);
                range.setStart(node, idx);
                range.setEnd(node, idx + query.length);
                const mark = document.createElement('mark');
                mark.className = 'oc-search-highlight';
                try { range.surroundContents(mark); } catch (_) {}
                searchResults.push(mark);
            }
        }
    });

    if (searchResults.length > 0) {
        document.getElementById('ocSearchCount').textContent = `1/${searchResults.length}`;
        document.getElementById('ocSearchPrev').disabled = false;
        document.getElementById('ocSearchNext').disabled = false;
        navigateSearch(1);
    } else {
        document.getElementById('ocSearchCount').textContent = '无匹配';
    }
}

function navigateSearch(dir) {
    searchResults.forEach(r => r.classList.remove('oc-search-active'));
    searchIndex = (searchIndex + dir + searchResults.length) % searchResults.length;
    const current = searchResults[searchIndex];
    current.classList.add('oc-search-active');
    current.scrollIntoView({ behavior: 'smooth', block: 'center' });
    document.getElementById('ocSearchCount').textContent = `${searchIndex + 1}/${searchResults.length}`;
}

function clearHighlights() {
    document.querySelectorAll('.oc-search-highlight').forEach(m => {
        const parent = m.parentNode;
        while (m.firstChild) parent.insertBefore(m.firstChild, m);
        parent.removeChild(m);
    });
}

function closeSearch() {
    clearHighlights();
    document.getElementById('ocSearchBar').style.display = 'none';
    document.getElementById('ocSearchInput').value = '';
    searchResults = [];
}
```

- [ ] **Step 3: 在启动时调用 initSearch()**

在 `chat.js` 的 `checkWebStatus()` 末尾或 `DOMContentLoaded` 回调中，调用 `initSearch()`。

- [ ] **Step 4: 在 style.css 中添加搜索样式**

```css
/* ======== 搜索栏 ======== */
.oc-search-bar { display: flex; align-items: center; gap: 8px; padding: 4px 12px; background: var(--bg-panel); border-bottom: 1px solid var(--border); }
.oc-search-bar input { flex: 1; padding: 4px 8px; border: 1px solid var(--border); border-radius: 4px; background: var(--bg-input); color: var(--text); font-size: 13px; outline: none; }
.oc-search-bar input:focus { border-color: var(--accent); }
.oc-search-count { font-size: 11px; color: var(--text-dim); min-width: 60px; text-align: center; }
.oc-search-highlight { background: var(--search-highlight, #ffeb3b); color: #000; border-radius: 2px; padding: 0 1px; }
.oc-search-highlight.oc-search-active { background: var(--accent); color: #fff; }
```

- [ ] **Step 5: 编译验证**

```bash
go build ./...
```

- [ ] **Step 6: 提交**

```cmd
git add frontend/dist/index.html frontend/dist/chat.js frontend/dist/style.css
git commit -m "feat: 工作区增加 Ctrl+F 消息搜索"
```

---

### Task 8: 清理与最终验证

- [ ] **Step 1: 清理 index.html 中的残留引用**

搜索 `index.html` 中所有 `view-models` 引用，确认已全部改为 `view-omo`。

- [ ] **Step 2: 清理 main.js 中的残留引用**

搜索 `main.js` 中所有 `view-models`、`model-config`、`BatchResult`、`TargetInfo` 引用，确认无残留。

- [ ] **Step 3: 清理 navigation.js**

确认 `navigation.js` 中无 `view-models` 引用。

- [ ] **Step 4: 运行完整编译**

```bash
go build ./...
```

预期: 零错误。

- [ ] **Step 5: 运行完整测试**

```bash
go test ./...
```

预期: 所有测试通过。如果 `skill/skill.go` 重构导致 `config_test.go` 中引用了旧类型，需要同步更新测试。

- [ ] **Step 6: 运行 lsp_diagnostics**

对改动的 Go 文件运行 LSP 诊断，确保无新增警告。

- [ ] **Step 7: 最终提交**

```cmd
git add -A
git commit -m "chore: 清理残留引用，最终验证通过"
```
