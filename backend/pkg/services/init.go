package services

import (
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/services/market"
)

func LoadServices(cfgs *common.Services) error {
	if cfgs == nil {
		return nil
	}
	err := market.LoadConfig(cfgs.McpMarket)
	if err != nil {
		return fmt.Errorf("load market config failed: %w", err)
	}
	return nil
}
