//go:build !enterprise

package app

import "github.com/gin-gonic/gin"

// enterpriseTableInitializers returns an empty slice when enterprise features are disabled.
func enterpriseTableInitializers() []func() (string, error) {
	return nil
}

// registerEnterprisePlugin is a no-op when enterprise features are disabled.
func registerEnterprisePlugin() error {
	return nil
}

// registerEnterpriseRoutes is a no-op when enterprise features are disabled.
func registerEnterpriseRoutes(engine *gin.Engine, routerPrefix string) {
}
