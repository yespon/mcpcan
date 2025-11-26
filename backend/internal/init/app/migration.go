package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

// RunMigrations 执行数据库迁移，特别是将 tokens 从 JSON 字段迁移到独立的表中
func RunMigrations(db *gorm.DB) {
	// 1. 自动创建/更新表结构以确保所有表和列都存在
	fmt.Println("Running AutoMigrate...")
	if err := db.AutoMigrate(&model.McpInstance{}, &model.McpToken{}, &model.Migration{}); err != nil {
		fmt.Printf("AutoMigrate failed: %v\n", err)
		// 在无法更新表结构时停止，因为后续操作会失败
		return
	}
	fmt.Println("AutoMigrate completed.")

	// 2. 检查迁移任务是否已经完成，实现幂等性
	const migrationName = "migrate-tokens-from-json-to-table"
	var migration model.Migration
	result := db.Where("name = ?", migrationName).First(&migration)

	if result.Error == nil {
		fmt.Printf("Migration '%s' has already been completed. Skipping.\n", migrationName)
		return
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Printf("Error checking migration status: %v\n", result.Error)
		return
	}

	// --- 开始迁移过程 ---
	fmt.Printf("Starting data migration: %s\n", migrationName)

	// 3. 查找所有需要迁移的旧数据
	// 定义一个只包含旧 tokens 字段的临时结构体，以避免加载关联数据
	type McpInstanceWithOldTokens struct {
		ID     uint            `gorm:"primarykey"`
		Tokens json.RawMessage `gorm:"type:json"`
	}

	var instancesToMigrate []McpInstanceWithOldTokens
	// 只选择那些 `tokens` 字段非空、非'[]'、非'null'的记录
	if err := db.Table("mcpcan_instance").Where("tokens IS NOT NULL AND tokens != '[]' AND tokens != '' AND tokens != 'null'").Find(&instancesToMigrate).Error; err != nil {
		fmt.Printf("Error fetching instances for migration: %v\n", err)
		return
	}

	if len(instancesToMigrate) == 0 {
		fmt.Println("No instances with old token data found. Migration not needed.")
	} else {
		fmt.Printf("Found %d instances to migrate.\n", len(instancesToMigrate))
		// 4. 遍历并迁移每个实例的数据
		for _, instance := range instancesToMigrate {
			// 旧的 McpToken 定义，用于反序列化
			type OldMcpToken struct {
				TokenType        model.TokenType   `json:"tokenType"`
				Token            string            `json:"token"`
				Headers          map[string]string `json:"headers,omitempty"`
				EnabledTransport bool              `json:"enabledTransport"`
				ExpireAt         int64             `json:"expireAt"`
				PublishAt        int64             `json:"publishAt"`
				Usages           []string          `json:"usages"`
			}

			var oldTokens []OldMcpToken
			if err := json.Unmarshal(instance.Tokens, &oldTokens); err != nil {
				fmt.Printf("[WARN] Could not unmarshal tokens for instance ID %d: %v. Skipping.\n", instance.ID, err)
				continue
			}

			if len(oldTokens) == 0 {
				continue
			}

			var newTokensToCreate []model.McpCanTokens
			for _, t := range oldTokens {
				headersJSON, _ := json.Marshal(t.Headers)
				usagesJSON, _ := json.Marshal(t.Usages)

				newTokensToCreate = append(newTokensToCreate, model.McpCanTokens{
					InstanceID:       instance.ID, // 关键：设置外键
					TokenType:        t.TokenType,
					Token:            t.Token,
					EnabledTransport: t.EnabledTransport,
					Headers:          json.RawMessage(headersJSON),
					Usages:           json.RawMessage(usagesJSON),
					ExpireAt:         t.ExpireAt,
					PublishAt:        t.PublishAt,
				})
			}

			// 5. 在事务中创建新记录并清空旧字段
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(&newTokensToCreate).Error; err != nil {
					// 如果令牌已经存在（uniqueIndex冲突），这可能不是一个致命错误，但需要记录
					fmt.Printf("[WARN] Could not create new tokens for instance ID %d (might be duplicates): %v\n", instance.ID, err)
					// 即使创建失败，我们仍然尝试清空旧字段，因为数据可能已部分迁移
				}

				// 清空旧的 JSON 字段以防重复迁移
				if err := tx.Table("mcpcan_instance").Where("id = ?", instance.ID).Update("tokens", "[]").Error; err != nil {
					return fmt.Errorf("failed to clear old tokens field for instance ID %d: %w", instance.ID, err)
				}
				return nil
			})

			if err != nil {
				fmt.Printf("[ERROR] Failed to migrate tokens for instance ID %d: %v\n", instance.ID, err)
			} else {
				fmt.Printf("Successfully migrated %d tokens for instance ID %d.\n", len(newTokensToCreate), instance.ID)
			}
		}
	}

	// 6. 记录迁移完成
	completedMigration := model.Migration{
		Name:        migrationName,
		CompletedAt: time.Now(),
	}
	if err := db.Create(&completedMigration).Error; err != nil {
		fmt.Printf("[ERROR] Failed to record migration completion for '%s': %v\n", migrationName, err)
	}

	fmt.Printf("Data migration '%s' finished.\n", migrationName)
}
