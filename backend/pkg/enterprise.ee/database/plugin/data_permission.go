package plugin

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/api/authz/user_auth"
	datapermissionpb "github.com/kymo-mcp/mcpcan/api/market/data_permission"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/gomap"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DataPermissionTableWhiteList maps database table names to data_type in permission queries.
var DataPermissionTableWhiteList = map[string]string{
	"mcpcan_instance":           "instance",
	"mcpcan_openapi_package":    "openapi_package",
	"mcpcan_code_package":       "code_package",
	"mcpcan_tokens":             "instance_tokens",
	"mcpcan_intelligent_access": "intelligent_access",
}

type GlobalDataPermissionPlugin struct{}

func (p *GlobalDataPermissionPlugin) Name() string {
	return "GlobalDataPermissionPlugin"
}

func (p *GlobalDataPermissionPlugin) Initialize(db *gorm.DB) error {
	_ = db.Callback().Query().Before("gorm:query").Register("global_data_permission:query", p.queryCallback)
	_ = db.Callback().Create().After("gorm:create").Register("global_data_permission:create", p.createCallback)
	return nil
}

func (p *GlobalDataPermissionPlugin) createCallback(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}
	tableName := db.Statement.Table

	dataType, ok := DataPermissionTableWhiteList[tableName]
	if !ok {
		return
	}

	userInfoRaw := gomap.Get(common.UserInfoContextKey)
	if userInfoRaw == nil {
		return
	}

	userInfo, ok := userInfoRaw.(*user_auth.UserInfo)
	if !ok || userInfo == nil || userInfo.UserId == 0 {
		return
	}

	userID := userInfo.UserId

	idColumn := "id"
	switch tableName {
	case "mcpcan_instance":
		idColumn = "instance_id"
	case "mcpcan_intelligent_access":
		idColumn = "access_id"
	case "mcpcan_code_package":
		idColumn = "package_id"
	case "mcpcan_openapi_package":
		idColumn = "openapi_file_id"
	}

	var dataIDs []interface{}
	extractID := func(rv reflect.Value) {
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return
		}
		field := db.Statement.Schema.LookUpField(idColumn)
		var dataID interface{}
		if field != nil {
			dataID, _ = field.ValueOf(db.Statement.Context, rv)
		}
		if dataID == nil || dataID == 0 || dataID == int64(0) || dataID == "" {
			if db.Statement.Schema.PrioritizedPrimaryField != nil {
				dataID, _ = db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv)
			}
		}
		if dataID != nil && dataID != 0 && dataID != int64(0) && dataID != "" {
			dataIDs = append(dataIDs, dataID)
		}
	}

	rv := db.Statement.ReflectValue
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			extractID(rv.Index(i))
		}
	case reflect.Struct:
		extractID(rv)
	}

	if len(dataIDs) == 0 {
		return
	}

	now := time.Now()
	for _, dataID := range dataIDs {
		err := db.Session(&gorm.Session{NewDB: true}).Exec(`INSERT INTO sys_data_permission (data_type, data_id, target_type, target_id, is_blacklist, created_by, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			dataType, fmt.Sprintf("%v", dataID), datapermissionpb.TargetType_USER.String(), userID, 0, userID, now).Error
		if err != nil {
			logger.Error("Failed to insert sys_data_permission", zap.Error(err), zap.String("dataType", dataType), zap.Any("dataID", dataID))
		}
	}
}

func (p *GlobalDataPermissionPlugin) queryCallback(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}
	tableName := db.Statement.Table

	dataType, ok := DataPermissionTableWhiteList[tableName]
	if !ok {
		return
	}

	userInfoRaw := gomap.Get(common.UserInfoContextKey)
	if userInfoRaw == nil {
		return
	}

	userInfo, ok := userInfoRaw.(*user_auth.UserInfo)
	if !ok || userInfo == nil || userInfo.UserId == 0 {
		return
	}
	// 管理员用户不进行权限过滤
	if userInfo.Username == "admin" {
		return
	}

	userID := userInfo.UserId

	var deptIDs []int64
	if userInfo.DeptId != 0 {
		deptIDs = append(deptIDs, userInfo.DeptId)
	}

	var roleIDs []int64
	if len(userInfo.RoleIds) > 0 {
		roleIDs = userInfo.RoleIds
	}

	deptMarks := makeMarks(len(deptIDs))
	roleMarks := makeMarks(len(roleIDs))

	var args []interface{}
	args = append(args, dataType, userID)

	for _, id := range deptIDs {
		args = append(args, id)
	}
	for _, id := range roleIDs {
		args = append(args, id)
	}

	idColumn := "id"
	if tableName == "mcpcan_instance" {
		idColumn = "instance_id"
	} else if tableName == "mcpcan_intelligent_access" {
		idColumn = "access_id"
	} else if tableName == "mcpcan_code_package" {
		idColumn = "package_id"
	} else if tableName == "mcpcan_openapi_package" {
		idColumn = "openapi_file_id"
	}

	whitelistSQL := fmt.Sprintf(`
		%s.%s IN (
			SELECT data_id FROM sys_data_permission 
			WHERE data_type = ? 
			  AND is_blacklist = 0 
			  AND (
				target_type = '%s' 
				OR (target_type = '%s' AND target_id = ?) 
				OR (target_type = '%s' AND target_id IN (%s)) 
				OR (target_type = '%s' AND target_id IN (%s))
			  )
		)`, tableName, idColumn, datapermissionpb.TargetType_ALL.String(), datapermissionpb.TargetType_USER.String(), datapermissionpb.TargetType_DEPT.String(), deptMarks, datapermissionpb.TargetType_ROLE.String(), roleMarks)

	blacklistSQL := fmt.Sprintf(`
		%s.%s NOT IN (
			SELECT data_id FROM sys_data_permission 
			WHERE data_type = ? 
			  AND target_type = '%s' 
			  AND target_id = ? 
			  AND is_blacklist = 1
		)`, tableName, idColumn, datapermissionpb.TargetType_USER.String())

	args = append(args, dataType, userID)

	db.Where(whitelistSQL+" AND "+blacklistSQL, args...)
}

func makeMarks(n int) string {
	if n == 0 {
		return "''"
	}
	marks := make([]string, n)
	for i := range marks {
		marks[i] = "?"
	}
	return strings.Join(marks, ",")
}
