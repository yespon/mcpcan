package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type sseWriter struct {
	http.ResponseWriter
	prefix []byte
}

func (w *sseWriter) Write(b []byte) (int, error) {
	// 拦截包含 "data: /" 的响应体（通常是 SSE 的 endpoint 下发阶段）
	if bytes.Contains(b, []byte("data: /")) {
		lines := bytes.Split(b, []byte("\n"))
		for i, line := range lines {
			if bytes.HasPrefix(line, []byte("data: /")) {
				// 获取原始的相对路径，例如 "data: /messages?sessionId=..." -> "messages?sessionId=..."
				origPath := string(line[6:]) 
				// 替换为绝对路径： "data: /mcp-gateway/xxx/messages?sessionId=..."
				lines[i] = append(append([]byte("data: "), w.prefix...), []byte("/"+strings.TrimPrefix(origPath, "/"))...)
			}
		}
		replaced := bytes.Join(lines, []byte("\n"))
		_, err := w.ResponseWriter.Write(replaced)
		// ReverseProxy 的 io.Copy 要求写入的长度要与读出的相等，所以这里欺骗性地返回原长度 b
		return len(b), err
	}
	
	return w.ResponseWriter.Write(b)
}

func (w *sseWriter) WriteHeader(statusCode int) {
	// 因为我们改写了 body 长度，必须删除 Content-Length，以防客户端提前截断
	w.ResponseWriter.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *sseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func main() {
	targetURLStr := os.Getenv("MCP_TARGET_URL")
	if targetURLStr == "" {
		log.Fatal("Fatal: MCP_TARGET_URL is required (e.g. http://127.0.0.1:8080)")
	}
	
	prefixStr := os.Getenv("MCP_ROUTE_PREFIX")
	if prefixStr == "" {
		log.Printf("Warning: MCP_ROUTE_PREFIX is empty, endpoint rewriting is disabled")
	} else {
		// 规范化，确保前缀无结尾斜杠且有开头斜杠
		prefixStr = "/" + strings.Trim(prefixStr, "/")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	targetURL, err := url.Parse(targetURLStr)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	// 对于 SSE 至关重要：遇到数据立刻刷新，不缓冲
	proxy.FlushInterval = -1

	// Director 也是 ReverseProxy 自带的，确保 Host Header 等完美透传
	// 如果需要覆盖原 Director 可以自己写，默认实现已经做得很好了

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 判断是否需要拦截响应体：只对于请求含有 event-stream 并且设置了 prefix 的情况进行包装
		// Websocket 升级请求会自动由 ReverseProxy 接管，不受影响
		if prefixStr != "" && (strings.Contains(r.Header.Get("Accept"), "text/event-stream") || strings.Contains(r.URL.Path, "sse")) {
			sw := &sseWriter{
				ResponseWriter: w,
				prefix:         []byte(prefixStr),
			}
			proxy.ServeHTTP(sw, r)
		} else {
			proxy.ServeHTTP(w, r)
		}
	})

	log.Printf("Starting MCP Transparency Proxy Sidecar on :%s", port)
	log.Printf("Target Upstream: %s", targetURLStr)
	log.Printf("Route Prefix (Rewrite): %s", prefixStr)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
