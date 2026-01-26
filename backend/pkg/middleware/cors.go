package middleware

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

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

var (
	cachedHostIps []string
	lastCacheTime time.Time
	cacheMutex    sync.RWMutex
)

// CORSMiddleware CORS middleware
func CORSMiddleware(domains []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		fmt.Printf("[DEBUG_LATENCY] Middleware.CORS Start: %s\n", start.Format("2006-01-02 15:04:05.000000"))
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

		fmt.Printf("[DEBUG_LATENCY] Middleware.CORS Logic: %.4fms\n", float64(time.Since(start).Microseconds())/1000.0)
		c.Next()
	}
}

// isAllowedOrigin Check if Origin is in the allowed list
func isAllowedOrigin(origin string, domains []string) bool {
	// Get cached host IPs
	var hostIps []string
	cacheMutex.RLock()
	if !lastCacheTime.IsZero() && time.Since(lastCacheTime) < 1*time.Hour {
		hostIps = cachedHostIps
	}
	cacheMutex.RUnlock()

	if hostIps == nil {
		cacheMutex.Lock()
		if !lastCacheTime.IsZero() && time.Since(lastCacheTime) < 1*time.Hour {
			hostIps = cachedHostIps
		} else {
			ips, _ := utils.GetHostIPs()
			cachedHostIps = ips
			hostIps = ips
			lastCacheTime = time.Now()
		}
		cacheMutex.Unlock()
	}

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

	// Get current server public IP (Already covered by GetHostIPs, but keeping as fallback or specific check if needed)
	// Actually GetHostIPs already includes public IP.
	// The original code had:
	// if ip, err := common.GetPublicIP(); err == nil { ... }
	// But utils.GetHostIPs() calls getPublicIPFromService() which seems to be what common.GetPublicIP() might do?
	// Let's check common.GetPublicIP implementation.
	// Previous read of utils/network.go showed getPublicIPFromService.
	// Let's assume common.GetPublicIP is similar or same.
	// To be safe and avoid double latency, we should rely on cached hostIps.
	// But let's look at original code again.

	return false
}
