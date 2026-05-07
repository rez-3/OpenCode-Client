# OC Manager 功能增强 — 设计规格

> **日期**: 2026-05-07
> **范围**: 三项核心改动，在现有架构上增量优化，不动工作区核心逻辑

---

## 一、改动概览

| # | 模块 | 改动类型 | 说明 |
|---|------|---------|------|
| 1 | 工作区 | 体验优化 | 会话 Tab 标签 + 面板拖拽 + 会话搜索 |
| 2 | 模型配置 | 文案修正 | 改名为「OMO 配置」 |
| 3 | 技能管理 | 重构 | 直接读 opencode 技能目录 + 预览/编辑 SKILL.md + 去多平台 |
| 4 | 配置文件 | 兼容修复 | `.jsonc` 不存在时 fallback 到 `.json` |

---

## 二、改动 1：工作区交互体验优化

### 2.1 会话 Tab 标签

**位置**: 消息区顶部（`oc-chat-head` 区域）

**行为**:
- 项目树中点击会话 → 打开新 Tab（若已存在则切换到该 Tab）
- Tab 显示会话标题（截断 18 字符），右侧 `×` 关闭按钮
- 点击 `×` → 关闭 Tab，切换到右边最近的 Tab；若无则显示"未选择会话"
- 最多同时打开 8 个 Tab，超出时关闭最旧未激活的 Tab
- Tab 切换时不销毁 DOM，用 `display:none/block` 控制可见性，保留 `messageCache` 缓存

**DOM 结构**:
```html
<div class="oc-tabs" id="ocTabs"><!-- 动态渲染 Tab 标签 --></div>
```

**数据模型**（JS 全局变量）:
```js
let openTabs = [];  // [{ id, title, active }]
```

### 2.2 左右面板可拖拽调整

**位置**:
- 项目树面板右侧边缘
- 右侧信息面板左侧边缘

**交互**:
- 拖拽手柄宽 4px，默认透明，hover 时显示蓝色竖线
- 鼠标按下后进入拖拽模式，`mousemove` 实时更新面板宽度
- 释放时宽度写入 `localStorage('oc-panel-widths')`，下次启动恢复
- 最小宽度：项目树 160px，右侧面板 200px
- 拖拽手柄 `cursor: col-resize`

**CSS 关键**:
```css
.oc-resize-handle { width: 4px; cursor: col-resize; }
.oc-resize-handle:hover, .oc-resize-handle.active { background: var(--accent); }
```

### 2.3 会话区搜索

**触发**: `Ctrl+F` 或点击搜索图标按钮

**位置**: 消息区顶部搜索栏

**行为**:
- 弹出搜索输入框，输入即搜（防抖 200ms）
- 在当前激活 Tab 的所有消息文本中全文匹配（大小写不敏感）
- 命中项黄色高亮（`.oc-search-highlight`），显示"第 N/M 个匹配"
- 上下箭头按钮跳转到上一个/下一个匹配项
- `Esc` 或点击 `×` 关闭搜索，清除高亮
- 搜索自动展开折叠的 reasoning/tool 块

---

## 三、改动 2：模型配置 → OMO 配置

### 3.1 改名范围

| 位置 | 原文 | 改为 |
|------|------|------|
| 侧边栏导航 | `⚙️ 模型配置` | `⚙️ OMO 配置` |
| 导航 data-view | `view-models` | `view-omo`（同步修改 `index.html` 中的 `data-view` 属性） |
| 视图面板 ID | `view-models` | `view-omo` |
| navigation.js 路由 | `'view-models'` | `'view-omo'` |
| model-config.js 文件名 | `model-config.js` | 改名为 `omo-config.js`，`index.html` 中 `<script src>` 同步更新 |
| 页面标题 | 模型配置 | OMO 配置 |

### 3.2 不变的部分
- 配置文件路径：仍指向 `~/.config/opencode/oh-my-openagent.jsonc`
- 内部功能逻辑（CRUD、批量设置、保存）完全不变
- 后端 API（`GetModelConfig`、`UpdateModels` 等）不改名，仅改前端展示

---

## 四、改动 3：技能管理重构

### 4.1 数据源变更

| 维度 | 旧方案 | 新方案 |
|------|--------|--------|
| 源目录 | `~/.cc-switch/skills/` | opencode 技能目录 |
| 全局路径 | — | `~/.config/opencode/skills/` |
| 项目路径 | — | `<项目根>/.opencode/skills/`（若存在）。项目根由前端传入，初始为空（仅显示全局技能），用户可在技能管理页手动选择项目目录 |
| 扫描方式 | `skill.go:GetAllSkills()` 扫描目录 | 直接遍历上述两个目录的子目录 |
| 降级 | — | 若项目目录不存在 `.opencode/skills/`，仅显示全局技能，不报错 |
| 平台 | OpenCode / Claude Code / Codex 三目标 | 仅 OpenCode |

### 4.2 后端改动

**skill/skill.go 重构**:

```go
// Manager 改为基于 opencode 路径
type Manager struct {
    globalDir  string  // ~/.config/opencode/skills/
    projectDir string  // 动态设置的项目技能目录
}

func NewManager() *Manager {
    homeDir, _ := os.UserHomeDir()
    return &Manager{
        globalDir: filepath.Join(homeDir, ".config", "opencode", "skills"),
    }
}

// SetProjectDir 设置项目技能目录
func (m *Manager) SetProjectDir(dir string) {
    m.projectDir = filepath.Join(dir, ".opencode", "skills")
}
```

**app.go 绑定变更**:
- 移除 `GetTargets()`、`ToggleAllSkills()`（不再有 Claude/Codex 目标）
- 新增 `GetProjectSkills(projectDir string)` 获取项目级技能
- 新增 `ReadSkillContent(skillPath string) string` 读取 SKILL.md 内容
- 新增 `SaveSkillContent(skillPath string, content string) error` 保存 SKILL.md

### 4.3 前端改动

**skill-manager.js 重构**:

- 移除 Claude/Codex 平台列
- 技能列表显示：名称 | 描述 | 路径 | 操作（预览/编辑/开关）
- **预览**: 点击技能名 → Modal 弹窗，`marked.parse()` 渲染 SKILL.md
- **编辑流程**: 预览 Modal 中有「编辑」按钮 → 预览区切换为 `<textarea>`（等宽字体，80 字符宽，高度自适应内容） → 用户编辑 Markdown 源码 → 「保存」调用 `api.SaveSkillContent(path, content)` 写入文件 → 成功后 toast 提示并切回预览模式 → 「取消」放弃编辑切回预览
- **开关**: 保留 Toggle 启用/禁用符号链接

**统计栏**: 仅显示「全局技能 N 个」+ 「项目技能 M 个」

### 4.4 model/types.go 变更

移除不再需要的类型：
- `TargetInfo` — 不再有 Claude/Codex 目标
- `Stats` 中的 `TargetStats` — 不再按平台统计
- `BatchResult` — 不再有跨平台批量操作

新增类型：
```go
type SkillContent struct {
    Path    string `json:"path"`    // 技能目录路径
    Content string `json:"content"` // SKILL.md 文件内容
}
```

---

## 五、改动 4：配置文件扩展名兼容

### 5.1 问题

`config/model_config.go` 和 `config/provider_config.go` 硬编码 `.jsonc` 扩展名：

```go
// 当前：只认 .jsonc
return filepath.Join(home, ".config", "opencode", "oh-my-openagent.jsonc")
return filepath.Join(home, ".config", "opencode", "opencode.jsonc")
```

若用户配置为 `.json` 格式（无注释），上述路径找不到文件，加载失败。

### 5.2 修复

两处 `ConfigPath` / `OpenCodeConfigPath` 增加 fallback：

```go
func resolvePath(jsoncPath string) string {
    if _, err := os.Stat(jsoncPath); err == nil {
        return jsoncPath
    }
    jsonPath := strings.TrimSuffix(jsoncPath, ".jsonc") + ".json"
    if _, err := os.Stat(jsonPath); err == nil {
        return jsonPath
    }
    return jsoncPath // 都不存在时返回 .jsonc 路径，让后续逻辑报错处理
}
```

`ConfigPath()` 和 `OpenCodeConfigPath()` 内调用 `resolvePath()` 后返回实际存在的路径。

### 5.3 影响范围

| 文件 | 改动 |
|------|------|
| `config/model_config.go` | `ConfigPath()` 内调用 resolvePath，新增 resolvePath 函数 |
| `config/provider_config.go` | `OpenCodeConfigPath()` 内调用 resolvePath |
| 其余 | 不变（`SaveConfig`、`SaveProvider` 等已有方法自动使用 `ConfigPath()` 的返回值） |

---

## 六、影响分析

### 6.1 文件改动清单

| 文件 | 改动量 | 改动类型 |
|------|--------|---------|
| `frontend/dist/index.html` | ~30 行 | 新增 Tab 栏 + 搜索栏 + 拖拽手柄 HTML |
| `frontend/dist/chat.js` | ~80 行 | Tab 管理 + 搜索逻辑 + 拖拽逻辑 |
| `frontend/dist/navigation.js` | ~5 行 | 改名 view-omo + 路由 |
| `frontend/dist/skill-manager.js` | 重写 | 新数据源 + 预览/编辑 Modal |
| `frontend/dist/model-config.js` | 文件名改 `omo-config.js` | 内部文案改名 |
| `frontend/dist/style.css` | ~60 行 | Tab 样式 + 搜索样式 + 拖拽手柄样式 |
| `skill/skill.go` | 重写 | 数据源改为 opencode 目录 |
| `model/types.go` | ~20 行 | 清理旧类型 + 新增 SkillContent |
| `config/model_config.go` | ~10 行 | ConfigPath() 增加 .json fallback |
| `config/provider_config.go` | ~8 行 | OpenCodeConfigPath() 增加 .json fallback |
| `app.go` | ~15 行 | 新增/移除绑定方法 |
| `main.go` | 不变 | 不变 |
| `service/` | 不变 | 不变 |

### 6.2 不碰的部分
- 工作区核心消息渲染逻辑（`chat.js` 中的 `renderPart`、`handleOcEvent` 等）
- SSE 事件流
- 进程管理
- 供应商配置
- 常用命令视图

---

## 七、验收标准

### 改动 1：工作区体验
- [ ] 点击项目树会话 → 顶部出现 Tab 标签
- [ ] 可同时打开多个 Tab，切换时消息不丢失
- [ ] × 关闭 Tab，切换到相邻 Tab
- [ ] 项目树和右侧面板可拖拽调整宽度，重启后记忆
- [ ] Ctrl+F 弹出搜索，键入即搜，命中高亮，上下跳转

### 改动 2：OMO 配置
- [ ] 侧边栏显示「OMO 配置」
- [ ] 页面标题显示「OMO 配置」
- [ ] 原有模型配置功能全部正常

### 改动 3：技能管理
- [ ] 技能列表显示 opencode 全局 + 项目技能
- [ ] 不再显示 Claude/Codex 列
- [ ] 点击技能名 → Modal 预览 SKILL.md（Markdown 渲染）
- [ ] Modal 可切换到编辑模式 → 保存写入文件
- [ ] Toggle 开关启用/禁用正常

### 改动 4：配置文件兼容
- [ ] `.jsonc` 文件存在时正常读取
- [ ] `.jsonc` 不存在但 `.json` 存在时自动 fallback
- [ ] 两者都不存在时不崩溃（报明确错误）

### 整体
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
- [ ] Wails 运行无报错
