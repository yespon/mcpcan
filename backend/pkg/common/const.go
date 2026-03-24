package common

import "time"

// Global configuration constants
const (
	RunModeKymo = "kymo"
	RunModeDev  = "demo"
	RunModeProd = "prod"

	DefaultPageSize = 10
	MaxPageSize     = 100

	AccessTokenExpireTime  = 60 * 60 * 24     // 1 day
	RefreshTokenExpireTime = 60 * 60 * 24 * 7 // 7 days

	DefaultLanguage    = "en-US"
	DefaultTheme       = "light"
	AutoLogoutTime     = 30 * 60 // 30 minutes
	EnableNotification = true

	// Replay attack time window
	ReplayWindow = 5 * time.Second
	// Whether to enable anti-replay protection
	EnableReplay = false
	// Whether to enable anti-tampering protection
	EnableSign = false

	// Password strength validation configuration
	PasswordMinLength      = 6     // Minimum length
	PasswordMaxLength      = 128   // Maximum length
	PasswordRequireLetter  = true  // Whether to require letters
	PasswordRequireDigit   = true  // Whether to require digits
	PasswordRequireSpecial = false // Whether to require special characters (recommended but not enforced)
	PasswordMinASCII       = 32    // Minimum printable ASCII character value
	PasswordMaxASCII       = 126   // Maximum printable ASCII character value

	// Avatar upload path
	AvatarPath = "/avatar"
	// Image upload path
	ImagesPath = "/images"
	// Static resource access path prefix
	StaticPrefix = "/static"

	// Default hosting image address
	DefatuleHostingImg = "mcp-hosting"

	SourceServerName = "mcpcan"

	McpProxyServiceName = "mcp-gateway-svc"

	MarketServerPrefix = "MCP_MARKET_SERVER_PREFIX"

	MarketRoutePrefix = "/market"

	AuthzServerPrefix = "MCP_AUTHZ_SERVER_PREFIX"

	AuthzRoutePrefix = "/authz"

	GatewayServerPrefix = "MCP_GATEWAY_SERVER_PREFIX"
	GatewayRoutePrefix  = "/mcp-gateway"

	SidecarServerPortEnv = "MCP_SIDECAR_SERVER_PORT"
	HostingServerPortEnv = "MCP_HOSTING_SERVER_PORT"

	SidecarContainerSuffix = "-sidecar"

	EnvironmentDefaultName = "Default-Kubernetes-Env"

	UserInfoHeaderKey  = "X-Custom-User-Info"
	UserIdHeaderKey    = "X-Consum-User-Id"
	UserInfoContextKey = "userInfo"
)

// Global query range constraints
const (
	// DefaultQueryRange defines the default time window applied when no range is provided
	DefaultQueryRange = 7 * 24 * time.Hour
	// MaxQueryRange defines the maximum allowed query time window
	MaxQueryRange = 180 * 24 * time.Hour
)

var SupportImageTypes = []string{"jpg", "jpeg", "png", "gif", "webp"}
