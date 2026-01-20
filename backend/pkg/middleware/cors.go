package middleware

import (
	"net"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/utils"

	"github.com/gin-gonic/gin"
)

var defaultDomains = []string{
	"http://localhost",
	"http://127.0.0.1",
	"https://localhost",
	"https://127.0.0.1",
}

// CORSMiddleware CORS middleware
func CORSMiddleware(domains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// If no Origin header or Origin is empty, it's not a CORS request
		if origin == "" {
			c.Next()
			return
		}

		// Check if Origin is in the allowed list
		if isAllowedOrigin(origin, domains) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// If not in allowed list, return 403 Forbidden
			c.Header("Access-Control-Allow-Origin", "*")
			i18n.Forbidden(c, "CORS request rejected")
			c.Abort()
			return
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // Preflight request cached for 24 hours

		// Handle preflight request
		if c.Request.Method == "OPTIONS" {
			i18n.SuccessResponse(c, nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// isAllowedOrigin Check if Origin is in the allowed list
func isAllowedOrigin(origin string, domains []string) bool {
	hostIps, _ := utils.GetHostIPs()
	if len(hostIps) > 0 {
		for _, ip := range hostIps {
			if strings.HasPrefix(origin, ip) {
				return true
			}
		}
	}
	for _, domain := range domains {
		if strings.HasPrefix(origin, domain) {
			return true
		}
		if domain == "*" {
			return true
		}
	}
	// Check default allowed origins
	for _, allowedOrigin := range defaultDomains {
		if strings.HasPrefix(origin, allowedOrigin) {
			return true
		}
	}

	// Check if it is a local IP address
	host := strings.TrimPrefix(origin, "http://")
	host = strings.TrimPrefix(host, "https://")
	host = strings.Split(host, ":")[0] // Remove port

	// Check if it is a local IP address
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() {
			return true
		}
	}

	// Get current server public IP
	if ip, err := common.GetPublicIP(); err == nil {
		if strings.Contains(origin, ip) {
			return true
		}
	}

	return false
}
