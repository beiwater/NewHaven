package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func intFromAny(v any) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case uint64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(x)
		return n
	default:
		return 0
	}
}

func floatFromAny(v any) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case uint64:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	default:
		return 0
	}
}

func boolFromAny(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	default:
		return false
	}
}

func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func requestIDFromBody(r *http.Request) string {
	var body map[string]any
	if r.Body == nil {
		return ""
	}
	// Peek at requestId without consuming body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return ""
	}
	if rid, ok := body["requestId"].(string); ok {
		return rid
	}
	return ""
}
