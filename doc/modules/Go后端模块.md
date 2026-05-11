# Go 后端模块

> 返回 [项目架构深度分析](../项目架构深度分析.md)

---

## 一、模块定位

Go 后端是整个应用的**业务逻辑核心**和**前后端桥梁**，通过 Wails v2 框架将 Go 方法绑定到前端 JavaScript 运行时。

---

## 二、核心门面：app.go（~370行）

`App` 结构体是 Wails 绑定的**唯一入口**，所有前端调用的 Go 方法都定义在此。它本身不包含复杂业务逻辑，而是通过**依赖注入**协调各子模块。

### 2.1 结构体定义

```go
type App struct {
    ctx context.Context
    sm  *skill.Manager   // 技能管理器
}
```

### 2.2 方法分组

| 分组 | 方法 | 说明 |
|------|------|------|
| **技能管理** | `GetSkills()`, `ToggleSkill()`, `ReadSkillContent()`, `SaveSkillContent()`, `Refresh()`, `GetSourceDir()`, `GetStats()` | 委托 `skill.Manager` |
| **模型配置** | `GetModelConfig()`, `GetAgentDescriptions()`, `GetAvailableModels()`, `RefreshAvailableModels()`, `UpdateModels()`, `AddModelEntry()`, `DeleteModelEntry()`, `AddModelType()`, `DeleteModelType()`, `GetConfigPath()`, `GetFullConfig()`, `SaveFullConfig()` | 委托 `config` 包 + `models.go` |
| **方案管理** | `GetSchemeDir()`, `ListSchemes()`, `ReadScheme()`, `SaveScheme()`, `ExportConfig()`, `OpenSchemeDir()` | 委托 `config` 包 + `service/scheme.go` |
| **供应商配置** | `GetProviders()`, `SaveProvider()`, `DeleteProvider()`, `GetProviderConfigPath()` | 委托 `config` 包 |
| **Web 服务** | `StartOpenCodeWeb()`, `StopOpenCodeWeb()`, `GetWebStatus()`, `OpenCodeAPI()`, `CreateSession()`, `AnswerQuestion()`, `RejectQuestion()`, `GetProjectTree()`, `StartOpenCodeEvents()`, `StopOpenCodeEvents()`, `LaunchWindowsTerminal()`, `OpenDirectoryDialog()`, `GetOpenCodeCommands()`, `GetSessions()` | 委托 `service` 包 |
| **命令参考** | `GetCommands()` | 静态数据 |
| **通用** | `OpenDir()` | 跨平台文件管理器打开 |

### 2.3 类型别名

```go
// 类型定义已迁移至 model 子包，以下通过类型别名保持前端 bind 兼容
type (
    SkillInfo       = model.SkillInfo
    WebResult       = model.WebResult
    APIResult       = model.APIResult
    ModelEntry      = model.ModelEntry
    ProviderInfo    = model.ProviderInfo
    TreeNode        = model.TreeNode
    // ... 等
)
```

---

## 三、模型缓存：models.go（127行）

负责 AI 模型列表的**线程安全缓存**，双通道获取策略。

### 3.1 获取策略

```
HTTP 优先 → /provider API（快速）
    ↓ 失败
CLI 回退 → opencode models 命令（可靠）
```

### 3.2 线程安全

```go
var (
    cachedModels   []string
    cachedModelsMu sync.RWMutex
    cachedModelsOk bool
)
```

- 读：`RLock` / `RUnlock`
- 写：`Lock` / `Unlock`
- 首次加载为后台懒加载

---

## 四、数据契约：model/types.go（180行）

定义所有前后端通信的**共享数据结构**，确保 JSON 序列化/反序列化的一致性。

| 结构体 | 用途 |
|--------|------|
| `SkillInfo` | 技能信息（名称、描述、路径、链接状态） |
| `Stats` / `ToggleResult` | 技能统计和操作结果 |
| `OpenAgentConfig` | oh-my-openagent.jsonc 顶层模型配置 |
| `ModelEntry` / `ModelConfig` / `SchemeInfo` | 前端展示的模型条目和方案信息 |
| `OpenCodeConfig` / `ProviderEntry` / `ProviderInfo` | opencode.jsonc 供应商配置 |
| `WebResult` / `APIResult` / `ProxyConfig` | Web 服务和 API 透传 |
| `TreeNode` / `CmdPaletteItem` / `SessionInfo` | 项目树、命令面板、会话 |
| `CmdInfo` / `CmdGroup` | CLI/TUI 命令参考 |

---

## 五、命令参考：commands.go（85行）

纯静态数据，提供 CLI 和 TUI 两套命令参考，分为：

- **CLI 命令**：会话管理、代理管理、服务管理、配置管理、维护
- **TUI 命令**：会话管理（/new, /compact, /undo, /redo）、信息查看（/help, /models, /themes）、操作（/init, /connect, /editor, /export）

---

## 六、启动入口：main.go（37行）

```go
func main() {
    app := NewApp()
    err := wails.Run(&options.App{
        Title:     "OpenCode管理中心",
        Width:     1280, Height: 820,
        MinWidth:  960,  MinHeight: 640,
        AssetServer: &assetserver.Options{Assets: assets},
        OnStartup:  app.startup,
        OnDomReady: app.domReady,
        OnShutdown: app.shutdown,
        Bind: []interface{}{app},        // 唯一的绑定点
    })
}
```

- `//go:embed all:frontend/dist` 将前端静态资源嵌入二进制
- `Bind: []interface{}{app}` 使 App 所有公开方法可从前端调用
- `OnDomReady` 触发 `app-ready` 事件，前端据此开始初始化

---

## 七、模块间依赖关系

```
main.go
  └── app.go ── 创建 App → 协调所有子模块
       ├── skill.Manager (skill/skill.go)
       ├── config.LoadConfig / SaveConfig (config/model_config.go)
       ├── config.LoadAgentDescriptions / ApplyDescriptions (config/agent_descriptions.go)
       ├── config.GetProviders / SaveProvider (config/provider_config.go)
       ├── service.StartOpenCodeWeb / StopOpenCodeWeb (service/process.go)
       ├── service.OpenCodeAPI (service/api.go)
       ├── service.StartOpenCodeEvents (service/sse.go)
       ├── service.GetProjectTree (service/tree.go)
       ├── service.ListSchemes / SaveScheme / ExportConfig (service/scheme.go)
       └── models.go (getAvailableModels / fetchModels)
```

**关键设计**：`app.go` 是纯门面（Facade），不包含业务逻辑，所有实质操作委托给各子包。这保证了单一职责和各模块的可测试性。
