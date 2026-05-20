package service

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// 可动态切换根目录的文件服务器
type SwitchableFileServer struct {
	currentDir string
	handler    http.Handler
	mu         sync.RWMutex
}

func NewSwitchableFileServer(initialDir string) *SwitchableFileServer {
	sfs := &SwitchableFileServer{
		currentDir: initialDir,
	}
	sfs.updateHandler()
	return sfs
}

func (sfs *SwitchableFileServer) updateHandler() {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()
	sfs.handler = http.FileServer(http.Dir(sfs.currentDir))
}

// 动态修改文件目录
func (sfs *SwitchableFileServer) SetDirectory(dir string) error {
	cleanDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// 检查目录是否存在
	if _, err := os.Stat(cleanDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cleanDir, 0755); err != nil {
			return err
		}
		log.Printf("目录不存在，已创建: %s", cleanDir)
	}

	sfs.mu.Lock()
	sfs.currentDir = cleanDir
	sfs.mu.Unlock()

	// 关键：更新handler
	sfs.updateHandler()
	return nil
}

func (sfs *SwitchableFileServer) GetDirectory() string {
	sfs.mu.RLock()
	defer sfs.mu.RUnlock()
	return sfs.currentDir
}

func (sfs *SwitchableFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sfs.mu.RLock()
	handler := sfs.handler
	sfs.mu.RUnlock()

	// 使用 StripPrefix 处理路径
	http.StripPrefix("/files/", handler).ServeHTTP(w, r)
}
