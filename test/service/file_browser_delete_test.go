package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"oc-manager/service"
)

func TestDeleteBrowserEntryRemovesFile(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("写入测试文件失败: %v", err)
	}
	result, err := service.DeleteBrowserEntry(root, "/hello.txt")
	if err != nil {
		t.Fatalf("删除文件失败: %v", err)
	}
	if !result.Success {
		t.Fatalf("删除结果异常: %#v", result)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Fatalf("文件未被删除: %v", err)
	}
}

func TestDeleteBrowserEntryRemovesDirectory(t *testing.T) {
	root := t.TempDir()
	dirPath := filepath.Join(root, "subdir")
	if err := os.MkdirAll(filepath.Join(dirPath, "nested"), 0755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dirPath, "nested", "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatalf("写入目录内测试文件失败: %v", err)
	}
	result, err := service.DeleteBrowserEntry(root, "/subdir/")
	if err != nil {
		t.Fatalf("删除目录失败: %v", err)
	}
	if !result.Success {
		t.Fatalf("删除目录结果异常: %#v", result)
	}
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Fatalf("目录未被删除: %v", err)
	}
}
