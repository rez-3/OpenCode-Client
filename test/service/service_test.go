package service_test

import (
	"strings"
	"testing"

	"oc-manager/service"
)

func TestGetWebSessionNil(t *testing.T) {
	origHost := *service.TestLastCfgHost
	origPort := *service.TestLastCfgPort
	t.Cleanup(func() {
		*service.TestLastCfgHost = origHost
		*service.TestLastCfgPort = origPort
	})
	*service.TestLastCfgHost = service.TestDefaultHostname
	*service.TestLastCfgPort = 1

	service.TestWebSessMu.Lock()
	service.TestSetWebSessNil()
	service.TestWebSessMu.Unlock()
	if !service.TestGetWebSessionNilCheck() {
		t.Fatal("expected nil webSession when no serve is running")
	}
}

func TestBuildTreeJSONEmpty(t *testing.T) {
	result := service.TestBuildTreeJSONEmpty()
	if result != "[]" {
		t.Fatalf("empty tree should be [], got %s", result)
	}
}

func TestBuildTreeJSONGroups(t *testing.T) {
	projects := []service.ProjectInfo{
		{ID: "global", Name: "全局项目", Worktree: "/"},
	}
	sessions := []struct {
		ID, Title, ProjectID, Directory string
		UpdatedAt                       int64
	}{
		{ID: "ses_1", Title: "测试会话1", ProjectID: "global", Directory: "D:\\test"},
		{ID: "ses_2", Title: "测试会话2", ProjectID: "global", Directory: "D:\\test"},
		{ID: "ses_3", Title: "测试会话3", ProjectID: "global", Directory: "C:\\other"},
	}

	result := service.TestBuildTreeJSON(projects, sessions)

	for _, want := range []string{`"type":"project"`, `"type":"directory"`, `"type":"session"`, `ses_1`, `ses_2`, `ses_3`, `D:\\test`, `C:\\other`} {
		if !strings.Contains(result, want) {
			t.Fatalf("tree missing %q: %s", want, result)
		}
	}
}

func TestBuildTreeJSONIncludesUpdatedAt(t *testing.T) {
	projects := []service.ProjectInfo{
		{ID: "global", Name: "全局项目", Worktree: "/"},
	}
	sessions := []struct {
		ID, Title, ProjectID, Directory string
		UpdatedAt                       int64
	}{
		{ID: "ses_1", Title: "这是一个很长的会话标题需要在前端完整展示用于悬停气泡", ProjectID: "global", Directory: "D:\\test", UpdatedAt: 1746604200000},
	}

	result := service.TestBuildTreeJSON(projects, sessions)

	for _, want := range []string{`"title":"这是一个很长的会话标题需要在前端完整展示用于悬停气泡"`, `"updatedAt":"2025-05-07`} {
		if !strings.Contains(result, want) {
			t.Fatalf("tree missing %q: %s", want, result)
		}
	}
}
