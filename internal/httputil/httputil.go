// Package httputil 提供 HTTP handler 的通用工具函数。
package httputil

import (
	"encoding/json"
	"net/http"

	"oc-manager/model"
)

// GuardMethod 检查请求方法是否匹配，不匹配时写入 405 错误并返回 false。
func GuardMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// WriteJSON 将 v 序列化为 JSON 写入响应。
func WriteJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

// WriteError 将错误消息以 JSON 格式写入响应。
func WriteError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ErrNoService 返回 "opencode 服务未启动" 的标准错误。
func ErrNoService() model.APIResult {
	return model.APIResult{Error: "opencode 服务未启动"}
}
