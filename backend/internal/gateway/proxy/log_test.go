package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

func init() {
	if err := logger.Init("info", "json"); err != nil {
		panic(err)
	}
	// host: 134.175.7.229
	// port: 31306
	// database: mcp_dev
	// username: mcp_user
	// password: dev-password
	database.Init(&common.MySQLConfig{
		Host:     "134.175.7.229",
		Port:     31306,
		Username: "mcp_user",
		Password: "dev-password",
		Database: "mcp_dev",
	})
}

// TestGatewayLogQueueWriteToDB verifies writing via queue persists data in MySQL.
func TestGatewayLogQueueWriteToDB(t *testing.T) {
	token := fmt.Sprintf("tok-write-%d", time.Now().UnixNano())
	reqID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	payload := map[string]interface{}{
		"reqId": reqID,
		"index": 1,
	}
	extra, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed marshal extra: %v", err)
	}

	if err := RecordGatewayLog("inst-queue", model.TokenTypeBearer, token, 0, "queue-write", extra); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}

	// Poll the database until the record appears or timeout.
	ctx := context.Background()
	deadline := time.Now().Add(5 * time.Second)
	for {
		logs, qerr := mysql.GatewayLogRepo.FindByToken(ctx, token)
		if qerr == nil {
			for _, lg := range logs {
				if lg.Token == token && lg.Log == "queue-write" {
					// Basic assertions on stored fields.
					if lg.InstanceID != "inst-queue" {
						t.Fatalf("unexpected instanceID: %s", lg.InstanceID)
					}
					if lg.TokenType != model.TokenTypeBearer {
						t.Fatalf("unexpected tokenType: %s", lg.TokenType)
					}
					if len(lg.Extra) == 0 {
						t.Fatalf("extra is empty")
					}
					return
				}
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for persisted log for token=%s", token)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// fmtInt returns the decimal string of an int without allocations.
//
