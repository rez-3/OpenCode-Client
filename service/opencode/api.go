// Package service 处理 OpenCode serve 进程管理、API 代理、SSE 事件流、会话 CRUD、项目树构建和终端启动。
package opencode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"oc-manager/model"
)

// getWebSessionBase 返回 opencode serve 的基础 URL（http://host:port）。
func getWebSessionBase() (string, error) {
	sess := getWebSession()
	if sess == nil {
		return "", fmt.Errorf("opencode 服务未启动")
	}
	return fmt.Sprintf("http://%s:%d", sess.hostname, sess.port), nil
}

// OpenCodeAPI 代理访问本机 opencode serve API，避免前端跨域限制。
// body 为 JSON 对象时自动提取 "directory" 字段设为 x-opencode-directory 请求头。
func OpenCodeAPI(method, path, body string) model.APIResult {
	sess := getWebSession()
	if sess == nil {
		return model.APIResult{Error: "opencode 服务未启动"}
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := fmt.Sprintf("http://%s:%d%s", sess.hostname, sess.port, path)

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return model.APIResult{Error: err.Error()}
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.APIResult{Error: err.Error()}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.APIResult{Status: resp.StatusCode, Error: err.Error()}
	}
	return model.APIResult{Success: resp.StatusCode >= 200 && resp.StatusCode < 300, Status: resp.StatusCode, Body: string(data)}
}


// findQuestionID 从 /question API 查找匹配 sessionID 的待回答问题 ID。
func findQuestionID(base, sessionID string) (string, error) {
	resp, err := http.Get(base + "/question")
	if err != nil {
		return "", fmt.Errorf("获取问题列表失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	type QuestionRequest struct {
		ID        string `json:"id"`
		SessionID string `json:"sessionID"`
	}
	var questions []QuestionRequest
	if err := json.Unmarshal(body, &questions); err != nil {
		return "", fmt.Errorf("解析问题列表失败: %v", err)
	}

	for _, q := range questions {
		if q.SessionID == sessionID {
			return q.ID, nil
		}
	}
	return "", fmt.Errorf("未找到该会话的待回答问题")
}

// AnswerQuestion 回答 question 工具调用。
func AnswerQuestion(sessionID, answerLabel string) model.APIResult {
	base, err := getWebSessionBase()
	if err != nil {
		return model.APIResult{Error: err.Error()}
	}

	requestID, err := findQuestionID(base, sessionID)
	if err != nil {
		return model.APIResult{Error: err.Error()}
	}

	replyBody := fmt.Sprintf(`{"answers":[["%s"]]}`, answerLabel)
	replyResp, err := http.Post(
		fmt.Sprintf("%s/question/%s/reply", base, requestID),
		"application/json",
		strings.NewReader(replyBody),
	)
	if err != nil {
		return model.APIResult{Error: fmt.Sprintf("回答问题失败: %v", err)}
	}
	defer replyResp.Body.Close()

	replyData, _ := io.ReadAll(replyResp.Body)
	return model.APIResult{
		Success: replyResp.StatusCode >= 200 && replyResp.StatusCode < 300,
		Status:  replyResp.StatusCode,
		Body:    string(replyData),
	}
}

// RejectQuestion 忽略 question 工具调用。
func RejectQuestion(sessionID string) model.APIResult {
	base, err := getWebSessionBase()
	if err != nil {
		return model.APIResult{Error: err.Error()}
	}

	requestID, err := findQuestionID(base, sessionID)
	if err != nil {
		return model.APIResult{Error: err.Error()}
	}

	req, _ := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/question/%s/reject", base, requestID), nil)
	rejectResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.APIResult{Error: fmt.Sprintf("忽略问题失败: %v", err)}
	}
	defer rejectResp.Body.Close()

	rejectData, _ := io.ReadAll(rejectResp.Body)
	return model.APIResult{
		Success: rejectResp.StatusCode >= 200 && rejectResp.StatusCode < 300,
		Status:  rejectResp.StatusCode,
		Body:    string(rejectData),
	}
}

// ProjectInfo 项目树中的项目信息。
type ProjectInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Worktree string `json:"worktree"`
	VCS      string `json:"vcs"`
}

type sessionTime struct {
	Created int64 `json:"created"`
	Updated int64 `json:"updated"`
}

type treeSession struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	ProjectID string      `json:"projectID"`
	Directory string      `json:"directory"`
	Time      sessionTime `json:"time"`
}

// GetProjectTree 获取项目→目录→会话的树形结构 JSON。
// knownDirs 是前端记录的所有建过会话的目录（JSON 字符串数组），用于查询 global 项目会话。
func GetProjectTree(knownDirs string) string {
	base, err := getWebSessionBase()
	if err != nil {
		return "[]"
	}
	client := http.Client{Timeout: 10 * time.Second}

	var projects []ProjectInfo
	var extraDirs []string
	if knownDirs != "" {
		json.Unmarshal([]byte(knownDirs), &extraDirs)
	}

	// 获取项目列表
	resp1, err := client.Get(base + "/project")
	if err == nil {
		defer resp1.Body.Close()
		body, _ := io.ReadAll(resp1.Body)
		json.Unmarshal(body, &projects)
	} else {
		projects = []ProjectInfo{{ID: "global", Name: "全局项目", Worktree: "/"}}
	}

	var allSessions []treeSession
	seen := map[string]bool{}

	// 获取所有项目会话（all=true 会返回当前实例目录所属项目的全部会话）
	// 当前实例目录是 exe 所在目录，属于 git 项目，因此能获取 git 项目的所有会话
	resp, err := client.Get(base + "/session?all=true&roots=true&limit=500")
	if err == nil {
		defer resp.Body.Close()
		var batch []treeSession
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &batch)
		for _, s := range batch {
			if !seen[s.ID] {
				seen[s.ID] = true
				allSessions = append(allSessions, s)
			}
		}
	}

	// 查询已知 global 目录下的会话
	for _, dir := range extraDirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		resp, err := client.Get(base + "/session?directory=" + url.QueryEscape(dir) + "&roots=true&limit=200")
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		var batch []treeSession
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &batch)
		for _, s := range batch {
			if !seen[s.ID] {
				seen[s.ID] = true
				allSessions = append(allSessions, s)
			}
		}
	}

	return buildTreeJSON(projects, allSessions)
}

func buildTreeJSON(projects []ProjectInfo, sessions []treeSession) string {
	// 按 project 分组，再按 directory 分组
	projectMap := make(map[string]*model.TreeNode)
	dirMap := make(map[string]*model.TreeNode) // key: projectID+"|"+directory

	for _, p := range projects {
		name := p.Name
		if name == "" {
			name = p.ID
		}
		if name == "global" {
			name = "全局项目"
		}
		node := &model.TreeNode{ID: p.ID, Title: name, Type: "project"}
		projectMap[p.ID] = node
	}

	for _, s := range sessions {
		pid := s.ProjectID
		if pid == "" {
			pid = "global"
		}
		dir := s.Directory
		if dir == "" {
			continue
		}
		dirKey := pid + "|" + dir

		// 确保 project 存在
		proj, ok := projectMap[pid]
		if !ok {
			name := pid
			if pid == "global" {
				name = "全局项目"
			}
			proj = &model.TreeNode{ID: pid, Title: name, Type: "project"}
			projectMap[pid] = proj
		}

		// 确保 directory 节点存在
		dirNode, ok := dirMap[dirKey]
		if !ok {
			dirNode = &model.TreeNode{ID: dirKey, Title: dir, Type: "directory"}
			dirMap[dirKey] = dirNode
			proj.Children = append(proj.Children, *dirNode)
		}

		// 找到刚添加的 directory 节点引用
		title := s.Title
		if title == "" {
			title = s.ID
		}

		// 取第一个可用的时间字段（updated 优先，其次 created）
		var sessionTime string
		if s.Time.Updated > 0 {
			sessionTime = time.UnixMilli(s.Time.Updated).Format("2006-01-02 15:04")
		} else if s.Time.Created > 0 {
			sessionTime = time.UnixMilli(s.Time.Created).Format("2006-01-02 15:04")
		}
		for i := range proj.Children {
			if proj.Children[i].ID == dirKey {
				proj.Children[i].Children = append(proj.Children[i].Children, model.TreeNode{
					ID:        s.ID,
					Title:     title,
					Type:      "session",
					UpdatedAt: sessionTime,
					Directory: dir,
				})
			}
		}
	}

	// 转为数组
	tree := make([]model.TreeNode, 0, len(projectMap))
	for _, p := range projectMap {
		tree = append(tree, *p)
	}

	data, _ := json.Marshal(tree)
	return string(data)
}
