package service_test

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"oc-manager/service"
)

func TestUploadBrowserFileCreatesFile(t *testing.T) {
	root := t.TempDir()
	result, err := service.UploadBrowserFile(root, "/", "hello.txt", base64.StdEncoding.EncodeToString([]byte("hello world")), false)
	if err != nil {
		t.Fatalf("上传文件失败: %v", err)
	}
	if !result.Success || result.Conflict {
		t.Fatalf("上传结果异常: %#v", result)
	}
	data, err := os.ReadFile(filepath.Join(root, "hello.txt"))
	if err != nil {
		t.Fatalf("读取上传文件失败: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("上传文件内容异常: %q", string(data))
	}
}

func TestUploadBrowserFileDetectsConflict(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("写入初始文件失败: %v", err)
	}
	result, err := service.UploadBrowserFile(root, "/", "hello.txt", base64.StdEncoding.EncodeToString([]byte("new")), false)
	if err != nil {
		t.Fatalf("上传冲突文件失败: %v", err)
	}
	if result.Success || !result.Conflict {
		t.Fatalf("冲突结果异常: %#v", result)
	}
	data, err := os.ReadFile(filepath.Join(root, "hello.txt"))
	if err != nil {
		t.Fatalf("读取冲突文件失败: %v", err)
	}
	if string(data) != "old" {
		t.Fatalf("冲突情况下不应覆盖文件: %q", string(data))
	}
}

func TestUploadBrowserFileOverwritesWhenAllowed(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("old"), 0644); err != nil {
		t.Fatalf("写入初始文件失败: %v", err)
	}
	result, err := service.UploadBrowserFile(root, "/", "hello.txt", base64.StdEncoding.EncodeToString([]byte("new")), true)
	if err != nil {
		t.Fatalf("覆盖上传失败: %v", err)
	}
	if !result.Success || result.Conflict {
		t.Fatalf("覆盖上传结果异常: %#v", result)
	}
	data, err := os.ReadFile(filepath.Join(root, "hello.txt"))
	if err != nil {
		t.Fatalf("读取覆盖文件失败: %v", err)
	}
	if string(data) != "new" {
		t.Fatalf("覆盖后的文件内容异常: %q", string(data))
	}
}
