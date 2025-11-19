package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"

	"gorm.io/gorm"
)

// GatewayLogRepo is the global repository instance for gateway logs
var GatewayLogRepo *GatewayLogRepository

func init() {
	RegisterInit(func() {
		repo := NewGatewayLogRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_gateway_log table: %v", err))
		}
	})
}

// GatewayLogRepository encapsulates CRUD and query operations for mcp_gateway_log
type GatewayLogRepository struct{}

// NewGatewayLogRepository creates a new GatewayLogRepository instance
func NewGatewayLogRepository() *GatewayLogRepository {
	GatewayLogRepo = &GatewayLogRepository{}
	return GatewayLogRepo
}

// getDB returns the scoped gorm DB for GatewayLog model
func (r *GatewayLogRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.GatewayLog{})
}

// Create inserts a new gateway log record
func (r *GatewayLogRepository) Create(ctx context.Context, log *model.GatewayLog) error {
	now := time.Now()
	log.CreatedAt = now
	log.UpdatedAt = now
	return r.getDB().WithContext(ctx).Create(log).Error
}

// Update updates an existing gateway log record by ID
func (r *GatewayLogRepository) Update(ctx context.Context, log *model.GatewayLog) error {
	log.UpdatedAt = time.Now()
	return r.getDB().WithContext(ctx).Where("id = ?", log.ID).Updates(log).Error
}

// Delete removes a gateway log record by ID
func (r *GatewayLogRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB().WithContext(ctx).Where("id = ?", id).Delete(&model.GatewayLog{}).Error
}

// FindByID finds a gateway log by its primary ID
func (r *GatewayLogRepository) FindByID(ctx context.Context, id uint) (*model.GatewayLog, error) {
	var log model.GatewayLog
	if err := r.getDB().WithContext(ctx).Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// FindByInstanceID lists logs for a given instance ID, ordered by creation time desc
func (r *GatewayLogRepository) FindByInstanceID(ctx context.Context, instanceID string) ([]*model.GatewayLog, error) {
	var logs []*model.GatewayLog
	// Apply default last 24 hours time window
	start := time.Now().Add(-common.DefaultQueryRange)
	if err := r.getDB().WithContext(ctx).
		Where("instance_id = ?", instanceID).
		Where("created_at >= ?", start).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindByToken lists logs for a given token, ordered by creation time desc
func (r *GatewayLogRepository) FindByToken(ctx context.Context, token string) ([]*model.GatewayLog, error) {
	var logs []*model.GatewayLog
	// Apply default last 24 hours time window
	start := time.Now().Add(-common.DefaultQueryRange)
	if err := r.getDB().WithContext(ctx).
		Where("token = ?", token).
		Where("created_at >= ?", start).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindWithPagination paginates gateway logs with optional filters and sorting
// Supported filters: instance_id|string, instanceId|string, token|string, tokenType|model.TokenType,
// createdAtStart|time.Time, createdAtEnd|time.Time
func (r *GatewayLogRepository) FindWithPagination(
	ctx context.Context,
	page, pageSize int32,
	filters map[string]interface{},
	sortBy, sortOrder string,
) ([]*model.GatewayLog, int64, error) {
	var logs []*model.GatewayLog
	var total int64

	q := r.getDB().WithContext(ctx)

	var startTime, endTime time.Time
	var hasStart, hasEnd bool

	for key, v := range filters {
		switch key {
		case "instanceId":
			if s, ok := v.(string); ok && s != "" {
				q = q.Where("instance_id = ?", s)
			}
		case "token":
			if s, ok := v.(string); ok && s != "" {
				q = q.Where("token = ?", s)
			}
		case "tokenHeader":
			if s, ok := v.(string); ok && s != "" {
				q = q.Where("token_header = ?", s)
			}
		case "tokenType":
			if tt, ok := v.(model.TokenType); ok {
				q = q.Where("token_type = ?", tt)
			}
		case "event":
			if s, ok := v.(string); ok && s != "" {
				q = q.Where("event = ?", s)
			}
		case "level":
			if lv, ok := v.(int); ok {
				q = q.Where("level = ?", lv)
			}
		case "usages":
			if arr, ok := v.([]string); ok {
				vals := make([]string, 0, len(arr))
				for _, it := range arr {
					s := strings.TrimSpace(it)
					if s != "" {
						vals = append(vals, s)
					}
				}
				if len(vals) > 0 {
					placeholders := make([]string, 0, len(vals))
					for range vals {
						placeholders = append(placeholders, "FIND_IN_SET(?, usages)")
					}
					cond := strings.Join(placeholders, " OR ")
					args := make([]interface{}, 0, len(vals))
					for _, v := range vals {
						args = append(args, v)
					}
					q = q.Where(cond, args...)
				}
			}
		case "createdAtStart":
			if t, ok := v.(time.Time); ok && !t.IsZero() {
				startTime = t
				hasStart = true
			}
		case "createdAtEnd":
			if t, ok := v.(time.Time); ok && !t.IsZero() {
				endTime = t
				hasEnd = true
			}
		case "traceId":
			if s, ok := v.(string); ok && s != "" {
				q = q.Where("trace_id = ?", s)
			}
		}
	}

	// Apply default range if none provided; clamp and validate max range
	now := time.Now()
	if !hasStart && !hasEnd {
		startTime = now.Add(-common.DefaultQueryRange)
		endTime = now
		hasStart, hasEnd = true, true
	} else if hasStart && !hasEnd {
		endTime = now
		hasEnd = true
	} else if !hasStart && hasEnd {
		startTime = endTime.Add(-common.DefaultQueryRange)
		hasStart = true
	}

	// Ensure start <= end
	if hasStart && hasEnd && endTime.Before(startTime) {
		startTime, endTime = endTime, startTime
	}

	// Enforce maximum range window
	if hasStart && hasEnd {
		if endTime.Sub(startTime) > common.MaxQueryRange {
			return nil, 0, fmt.Errorf("query range exceeds max %d days", int(common.MaxQueryRange/(24*time.Hour)))
		}
		q = q.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	order := "DESC"
	if sortOrder == "asc" {
		order = "ASC"
	}

	switch sortBy {
	case "createdAt":
		q = q.Order(fmt.Sprintf("created_at %s", order))
	case "updatedAt":
		q = q.Order(fmt.Sprintf("updated_at %s", order))
	default:
		q = q.Order("created_at DESC")
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(int(offset)).Limit(int(pageSize)).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// InitTable migrates schema and ensures necessary indexes exist
func (r *GatewayLogRepository) InitTable() error {
	mod := &model.GatewayLog{}

	// Pre-fix legacy schema: ensure level column is INT with default 0
	var cnt struct{ Cnt int64 }
	tbl := mod.TableName()
	checkSQL := fmt.Sprintf("SELECT COUNT(*) AS cnt FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = '%v' AND column_name = 'level'", tbl)
	_ = r.getDB().Raw(checkSQL).Scan(&cnt).Error
	if cnt.Cnt > 0 {
		// Inspect column default and type
		typeInfo := struct {
			ColumnDefault *string
			DataType      string
		}{}
		inspectSQL := fmt.Sprintf("SELECT COLUMN_DEFAULT, DATA_TYPE FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = '%v' AND column_name = 'level' LIMIT 1", tbl)
		if err := r.getDB().Raw(inspectSQL).Scan(&typeInfo).Error; err == nil {
			needsAlter := false
			if typeInfo.DataType != "int" && typeInfo.DataType != "integer" && typeInfo.DataType != "bigint" && typeInfo.DataType != "mediumint" && typeInfo.DataType != "smallint" && typeInfo.DataType != "tinyint" {
				needsAlter = true
			}
			if typeInfo.ColumnDefault != nil && *typeInfo.ColumnDefault == "" {
				needsAlter = true
			}
			if needsAlter {
				alterSQL := fmt.Sprintf("ALTER TABLE %v MODIFY COLUMN level INT NOT NULL DEFAULT 0 COMMENT 'log level'", tbl)
				if err := r.getDB().Exec(alterSQL).Error; err != nil {
					return fmt.Errorf("failed to alter level column: %v", err)
				}
			}
		}
	}

	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table: %v", err)
	}

	var count int64

	// Index on instance_id for fast per-instance queries
	sql := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_instance_id'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_instance_id ON %v(instance_id)", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create instance_id index: %v", err)
		}
	}

	// Index on trace_id for fast per-trace queries
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_trace_id'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_trace_id ON %v(trace_id)", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create trace_id index: %v", err)
		}
	}

	// Index on token_header for filtering by authentication header source
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_token_header'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_token_header ON %v(token_header)", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create token_header index: %v", err)
		}
	}

	// Index on token for reverse lookups by token
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_token'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		// Use prefix index to avoid exceeding max key length (utf8mb4)
		sql2 := fmt.Sprintf("CREATE INDEX idx_token ON %v(token(100))", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create token index: %v", err)
		}
	}

	// Index on created_at for time range queries and ordering
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_created_at'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_created_at ON %v(created_at)", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create created_at index: %v", err)
		}
	}

	// Index on event for filtering by event type
	sql = fmt.Sprintf("SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = '%v' AND index_name = 'idx_event'", mod.TableName())
	r.getDB().Raw(sql).Count(&count)
	if count == 0 {
		sql2 := fmt.Sprintf("CREATE INDEX idx_event ON %v(event)", mod.TableName())
		if err := r.getDB().Exec(sql2).Error; err != nil {
			return fmt.Errorf("failed to create event index: %v", err)
		}
	}

	return nil
}
