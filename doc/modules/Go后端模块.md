# Go 后端模块

> 返回 [项目架构深度分析](../项目架构深度分析.md)

---

## 一、模块定位

Go 后端是应用的业务逻辑核心和前后端桥梁，通过 Wails v2 将 Go 方法绑定到前端 JavaScript。

---

## 二、核心门面：app.go（~370 行）

`App` 结构体是 Wails 绑定的唯一入口，所有前端调用的 Go 方法都定义在此。它本身不包含复杂业务逻辑，委托各子包执行。

### 2.1 结构体

```go
type App struct {
    ctx context.Context
    sm  *skill.Manager
}
```

### 2.2 方法分组

| 分组 | 方法 | 委托 |
|------|------|------|
| 技能管理 | `GetSkills`, `ToggleSkill`, `ReadSkillContent`, `SaveSkillContent`, `Refresh`, `GetSourceDir`, `GetStats` | `skill.Manager` |
| 模型配置 | `GetModelConfig`, `GetAgentDescriptions`, `GetAvailableModels`, `RefreshAvailableModels`, `UpdateModels`, `AddModelEntry`, `DeleteModelEntry`, `AddModelType`, `DeleteModelType`, `GetConfigPath`, `GetFullConfig`, `SaveFullConfig` | `config` 包 |
| 供应商配置 | `GetProviders`, `SaveProvider`, `DeleteProvider`, `GetProviderConfigPath` | `config` 包 |
| 方案管理 | `GetSchemeDir`, `ListSchemes`, `ReadScheme`, `SaveScheme`, `ExportConfig`, `OpenSchemeDir` | `config` + `service` 包 |
| Web 服务 | `StartOpenCodeWeb`, `StopOpenCodeWeb`, `GetWebStatus`, `OpenCodeAPI`, `CreateSession`, `AnswerQuestion`, `RejectQuestion`, `GetProjectTree`, `StartOpenCodeEvents`, `StopOpenCodeEvents`, `LaunchWindowsTerminal`, `OpenDirectoryDialog`, `GetOpenCodeCommands`, `GetSessions` | `service` 包 |
| 通用 | `OpenDir` | — |

其中 `GetProjectTree()` 负责返回项目 → 目录 → 会话的树形 JSON，`GetOpenCodeCommands()` 负责为 `/` 命令面板拉取动态命令列表；二者都委托给 `service` 层，不在 `app.go` 中内联业务逻辑。

其中与工作区最相关的链路是：

- `GetProjectTree()` → `service/tree.go` 构建项目 → 目录 → 会话树
- `StartOpenCodeEvents()` / `StopOpenCodeEvents()` → `service/sse.go` 转发 OpenCode 全局 SSE
- `GetOpenCodeCommands()` → `service/tree.go` 读取 `/command` 供 `/` 命令面板使用

### 2.3 类型别名

```go
type (
    SkillInfo    = model.SkillInfo
    ModelEntry   = model.ModelEntry
    ProviderInfo = model.ProviderInfo
    WebResult    = model.WebResult
    APIResult    = model.APIResult
    SchemeInfo   = model.SchemeInfo
    // ...
)
```

---

## 三、模型缓存：models.go（~130 行）

AI 模型列表的线程安全缓存，双通道获取：

```
HTTP 优先 → /provider API（快速）
    ↓ 失败
CLI 回退 → opencode models 命令（可靠）
```

```go
var (
    cachedModels   []string
    cachedModelsMu sync.RWMutex
    cachedModelsOk bool
)
```

---

## 四、数据契约：model/types.go

前后端共享数据结构：

| 结构体 | 用途 |
|--------|------|
| `ModelEntry` | 前端展示的模型条目（key, type, model, comment） |
| `ModelConfig` | JSONC 中的 model 配置块 |
| `OpenAgentConfig` | oh-my-openagent.jsonc 顶层结构 |
| `SchemeInfo` | 方案文件元信息（name, fileName, fullPath） |
| `ProviderInfo` | 供应商信息 |
| `SkillInfo` / `Stats` / `ToggleResult` | 技能信息和操作结果 |
| `WebResult` / `APIResult` | Web 服务状态和 API 响应 |
| `TreeNode` / `SessionInfo` | 项目树节点和会话信息 |
| `CmdInfo` / `CmdPaletteItem` | 命令参考 |

---

## 五、配置包：config/

| 文件 | 职责 |
|------|------|
| `model_config.go` | oh-my-openagent.jsonc 的行级 model 值替换。`LoadConfig()` 解析 JSONC，`SaveConfig()` 逐行替换 model 值，`ConfigToEntries()` 结合描述表生成 ModelEntry 列表 |
| `agent_descriptions.go` | `LoadAgentDescriptions()` 从 `exeDir/configs/oh-my-openagent/agents-comments.json` 加载描述表；`ApplyDescriptions()` 合并写入 |
| `provider_config.go` | opencode.jsonc 供应商配置的完整序列化读写 |
| `scheme_config.go` | 提供 `SchemeDir()` 和 `EnsureSchemeDir()` 路径工具 |

---

## 六、服务层：service/

| 文件 | 职责 |
|------|------|
| `process.go` | OpenCode serve 进程管理（启动/停止/健康检查/终端启动） |
| `api.go` | 通用 HTTP API 代理 + 会话创建 + question 应答 |
| `sse.go` | SSE 事件流透传到前端运行时事件 |
| `tree.go` | 项目树构建（项目→目录→会话三级结构）+ `/command` 命令面板数据 |
| `scheme.go` | 方案文件管理：`ListSchemes` / `ReadScheme` / `SaveScheme` / `ExportConfig` / `OpenSchemeDir` |

---

## 七、技能管理：skill/skill.go

符号链接扫描与管理：

- `GetAllSkills()` — 递归扫描技能目录，检测 SKILL.md
- `ToggleSkill(path, name, enable)` — 启用时 `mklink /J`（Windows）/ `os.Symlink`，禁用时 `os.Remove`
- `ReadSkillContent()` / `SaveSkillContent()` — 读写 SKILL.md

详见 [技能管理模块](技能管理模块.md)

---

## 八、模块间依赖

```
main.go
  └── app.go ── 创建 App → 协调所有子模块
       ├── skill.Manager
       ├── config.LoadConfig / SaveConfig
       ├── config.LoadAgentDescriptions / ApplyDescriptions
       ├── config.GetProviders / SaveProvider
       ├── service.StartOpenCodeWeb / StopOpenCodeWeb
       ├── service.OpenCodeAPI
       ├── service.StartOpenCodeEvents
       ├── service.GetProjectTree
       ├── service.ListSchemes / SaveScheme / ExportConfig
       └── models.go (模型缓存)
```

`app.go` 是纯门面（Facade），所有实质操作委托给子包，保证单一职责和可测试性。
