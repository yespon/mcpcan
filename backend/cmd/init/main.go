// @Deprecated
// 此部分逻辑已迁移至 market 服务中实现，已弃用。
// 验证通过后将清理。
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kymo-mcp/mcpcan/internal/init/app"
)

func main() {
	// 创建应用程序实例
	appInstance := app.New()

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

	fmt.Println("Admin user created successfully!")
}
