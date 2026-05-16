# 页面 Web 服务模块

> 返回 [技术方案](../技术方案.md)

---

## 一、模块定位

页面 Web 服务模块负责把桌面应用的前端页面、App 门面方法和 SSE 事件流暴露为可浏览器访问的 HTTP 服务。该模块位于 `service/frontend_web.go`，由 `app.go` 通过薄桥接暴露给前端按钮和 Wails 绑定。

---

## 二、核心职责

1. 启动/停止页面 Web 服务
2. 提供前端静态资源（`frontend/dist`）
3. 暴露 `/api/project-tree`、`/api/open-code`、`/api/app-call` 等 HTTP 接口
4. 提供 `/events` SSE 事件流
5. 维护页面 Web 服务状态（运行中 / 离线、URL、端口）

---

## 三、关键文件

| 文件 | 职责 |
|------|------|
| `service/frontend_web.go` | 页面 Web 服务宿主、路由、SSE 输出 |
| `app.go` | `StartFrontendWeb` / `StopFrontendWeb` / `GetFrontendWebStatus` / `AppCall` 桥接 |
| `test/frontend/frontend_web_test.go` | 页面 Web HTTP 路由与状态测试 |

---

## 四、服务模型

```go
type FrontendWebBridge interface {
    GetProjectTree(string) string
    OpenCodeAPI(string, string, string) model.APIResult
    CreateSession(string) model.APIResult
    GetAvailableModels() ([]string, error)
    StartOpenCodeEvents() model.APIResult
    StopOpenCodeEvents() model.APIResult
    StartOpenCodeWeb(int, string, model.ProxyConfig) model.WebResult
    GetWebStatus(string, int) model.WebResult
    StopOpenCodeWeb() model.WebResult
    AppCall(string, []json.RawMessage) (interface{}, error)
}
```

桥接接口保证页面 Web 服务位于 `service/`，而实际业务仍由 `App` 门面统一分发。

---

## 五、HTTP 路由

| 路由 | 说明 |
|------|------|
| `/` | 前端静态页面 |
| `/api/project-tree` | 项目树读取 |
| `/api/open-code` | OpenCode API 代理 |
| `/api/session/create` | 创建会话 |
| `/api/models` | 模型列表 |
| `/api/open-code-web/*` | OpenCode serve 启停与状态 |
| `/api/open-code-events/*` | OpenCode 事件流启停 |
| `/api/app-call` | 通用 RPC 分发 |
| `/events` | SSE 事件推送 |

---

## 六、测试策略

`test/frontend/frontend_web_test.go` 放在 `test/` 目录下，通过 stub bridge 直接验证：

- 首页静态资源可访问
- 页面 Web 服务启停与状态
- `app-call` 路由返回合法 JSON
- 目录树、技能浏览、技能保存等 Web 协议层行为

---

## 七、与其他模块的关系

- 目录浏览器见 [Web 目录浏览器模块](Web目录浏览器模块.md)
- 技能文件浏览/编辑见 [技能文件浏览编辑模块](技能文件浏览编辑模块.md)
- 工作区与移动端布局见 [工作区模块](工作区模块.md) 与 [手机端工作区模块](手机端工作区模块.md)
