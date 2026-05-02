# OC Manager

Wails v2 桌面应用，为 [OpenCode](https://github.com/nicepkg/opencode) 提供可视化管理界面。

## 一、项目概览

| 维度          | 详情                                                      |
| ------------- | --------------------------------------------------------- |
| **名称**      | OC Manager — OpenCode 可视化管理中心                      |
| **框架**      | Wails v2.12 (Go 后端 + WebView2 前端)                     |
| **语言**      | Go 1.25 + 原生 HTML/CSS/JS                                |
| **Go 源文件** | 11 个，全部在 `main` 包中                                 |
| **代码规模**  | Go ~2700 行 / JS ~3162 行 / CSS ~2500 行                  |
| **构建状态**  | ✅ `go build` 通过 / ✅ `go vet` 零警告 / ✅ 12 个测试全通过 |

---

## 二、架构分析

### 项目结构

```
skill-manager/
├── main.go              # Wails 入口，App 绑定到前端
├── app.go               # App 结构体（生命周期、技能管理、模型配置的门面方法）
├── skills.go            # SkillManager（技能目录扫描、符号链接 toggle）
├── symlink.go           # Windows 符号链接/联接（Junction）实现
├── web.go               # OpenCode Serve 进程管理 + API 代理 + SSE 事件流 + 会话树
├── config.go            # 模型配置 JSONC 读写（oh-my-openagent.jsonc）
├── provider.go          # 供应商配置读写（opencode.jsonc）
├── models.go            # 可用模型列表缓存（opencode models）
├── commands.go          # CLI/TUI 命令参考数据
├── config_test.go       # 模型配置测试（9个，覆盖良好）
├── web_test.go          # Web 可用性测试（3个，覆盖薄弱）
├── frontend/
│   ├── dist/
│   │   ├── index.html   # 页面结构
│   │   └── src/
│   │       ├── main.js  # 前端主逻辑（3162行，单文件）
│   │       ├── style.css # 样式表（VSCode 主题，双模式）
│   │       └── marked.min.js # Markdown 渲染
│   └── wailsjs/         # Wails 生成的 JS 绑定
├── build/
│   └── bin/             # 构建产物
├── go.mod / go.sum
├── wails.json
├── .gitignore
├── README.md
└── 技术方案.md
```

### 架构特点

所有逻辑集中在 `main` 包，没有分层（handler/service/repository）。对于桌面工具的规模来说尚可接受，但 `config.go`(756行) 和 `web.go`(738行) 已经过大。

### 文件职责关系图

```
main.go  ──启动──▶ App ──包含──▶ SkillManager
                     │
                     ├──▶ 技能管理 (skills.go + symlink.go)
                     ├──▶ 模型配置 (config.go)
                     ├──▶ 供应商配置 (provider.go)
                     ├──▶ 模型缓存 (models.go)
                     ├──▶ 服务管理 (web.go)
                     └──▶ 命令参考 (commands.go)
```

## 三、构建

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 构建
wails build

# 开发模式（热重载）
wails dev
```

构建产物：`build/bin/oc-manager.exe`

## 四、功能

- **服务管理** — 启动/停止 OpenCode Web Serve，自动检测已运行服务
- **会话管理** — 项目→目录→会话三级树，新建/切换/删除会话
- **消息查看** — Markdown 渲染、工具调用折叠、推理过程展示
- **实时推送** — SSE 事件流，流式输出增量更新
- **右侧面板** — 服务健康状态、代办事项、文件变更 diff
- **模型配置** — 按 agent/category 配置模型
- **技能管理** — 符号链接管理，多平台 toggle
- **常用命令** — CLI/TUI 命令参考
