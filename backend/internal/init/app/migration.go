// @Deprecated
// 此部分逻辑已迁移至 market 服务中实现，已弃用。
// 验证通过后将清理。
package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
)

// RunMigrations 执行数据库迁移，特别是将 tokens 从 JSON 字段迁移到独立的表中
func RunMigrations() {
	// 1. 自动创建/更新表结构以确保所有表和列都存在
	fmt.Println("Running AutoMigrate...")
	// 2. 检查迁移任务是否已经完成，实现幂等性
	const migrationName = "migrate-tokens-from-json-to-table"

	migration, _ := mysql.McpMigrationRepo.FindByName(context.Background(), migrationName)
	if migration != nil {
		fmt.Printf("Migration '%s' has already been completed. Skipping.\n", migrationName)
		return
	}

	// --- 开始迁移过程 ---
	fmt.Printf("Starting data migration: %s\n", migrationName)

	// 3. 查找所有需要迁移的旧数据
	// 定义一个只包含旧 tokens 字段的临时结构体，以避免加载关联数据
	instanceList, err := mysql.McpInstanceRepo.FindAll(context.Background())
	if err != nil {
		fmt.Printf("Error fetching instances for migration: %v\n", err)
		return
	}

	if len(instanceList) == 0 {
		fmt.Println("No instances with old token data found. Migration not needed.")
	} else {
		fmt.Printf("Found %d instances to migrate.\n", len(instanceList))
		// 4. 遍历并迁移每个实例的数据
		for _, instance := range instanceList {

			var oldTokens []model.McpToken
			if err := json.Unmarshal(instance.Tokens, &oldTokens); err != nil {
				fmt.Printf("[WARN] Could not unmarshal tokens for instance ID %d: %v. Skipping.\n", instance.ID, err)
				continue
			}

			if len(oldTokens) == 0 {
				continue
			}

			var newTokensToCreate []model.McpToken
			for _, t := range oldTokens {
				headersJSON, _ := json.Marshal(t.Headers)
				usagesJSON, _ := json.Marshal(t.Usages)

				newTokensToCreate = append(newTokensToCreate, model.McpToken{
					InstanceID: instance.InstanceID,
					Token:      t.Token,
					Enabled:    true,
					Headers:    json.RawMessage(headersJSON),
					Usages:     json.RawMessage(usagesJSON),
					ExpireAt:   t.ExpireAt,
					PublishAt:  t.PublishAt,
				})
			}

			if err := mysql.McpTokenRepo.CreateBatch(context.Background(), newTokensToCreate); err != nil {
				fmt.Printf("[WARN] Could not batch create tokens for instance ID %d: %v\n", instance.ID, err)
			}
		}
	}

	// 6. 记录迁移完成
	completedMigration := model.Migration{
		Name:        migrationName,
		CompletedAt: time.Now(),
	}
	if err := mysql.McpMigrationRepo.Create(context.Background(), &completedMigration); err != nil {
		fmt.Printf("[ERROR] Failed to record migration completion for '%s': %v\n", migrationName, err)
	}

	fmt.Printf("Data migration '%s' finished.\n", migrationName)
}
