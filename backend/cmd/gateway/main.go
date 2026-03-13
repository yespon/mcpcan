// Deprecated: 旧的网关启动入口，系统已迁移至 Traefik + Auth + Sidecar 模式。
// 此入口已不再使用，待新模式验证后将被清理。
package main

import (
	"log"
	"os"

	"github.com/kymo-mcp/mcpcan/internal/gateway/app"
)

// main 主函数
func main() {
	// 创建应用程序实例
	appInstance, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application instance: %v", err)
		os.Exit(1)
	}

	// 初始化应用程序
	if err := appInstance.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// 运行应用程序
	if err := appInstance.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
		os.Exit(1)
	}
}
