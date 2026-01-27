package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
)

// I18nMiddleware Internationalization middleware
func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get language code from request
		lang := GetLanguageFromRequest(c)

		// Parse supported language
		supportedLang := parseSupportedLanguage(lang)

		// Store language code in context
		i18nresp.SetLanguageToGin(c, supportedLang)

		// Set response header to inform client of current language
		c.Header("Accept-Language", string(supportedLang))

		c.Next()
	}
}

// parseSupportedLanguage Parse supported language
func parseSupportedLanguage(lang string) i18nresp.SupportedLanguage {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch {
	case strings.HasPrefix(lang, "zh"):
		return i18nresp.LanguageZhCN
	case strings.HasPrefix(lang, "en"):
		return i18nresp.LanguageEnUS
	default:
		return i18nresp.DefaultLanguage
	}
}

// GetLocalizer Get localizer from context (keep backward compatibility)
func GetLocalizer(c *gin.Context) interface{} {
	// This function keeps backward compatibility, but actually no longer uses go-i18n Localizer
	return nil
}

// GetMessage Get localized message from context
func GetMessage(c *gin.Context, messageID string, templateData map[string]interface{}) string {
	// This function is mainly for backward compatibility, actually uses pkg/i18n message system
	return messageID
}

// GetLanguage Get current language from context
func GetLanguage(c *gin.Context) string {
	lang := i18nresp.GetLanguageFromGin(c)
	return string(lang)
}

// LocalizedError Return localized error response
func LocalizedError(c *gin.Context, statusCode int, messageID string, templateData map[string]interface{}) {
	// Map status code to error code
	var errorCode int
	switch statusCode {
	case http.StatusBadRequest:
		errorCode = i18nresp.CodeBadRequest
	case http.StatusUnauthorized:
		errorCode = i18nresp.CodeUnauthorized
	case http.StatusForbidden:
		errorCode = i18nresp.CodeForbidden
	case http.StatusNotFound:
		errorCode = i18nresp.CodeNotFound
	case http.StatusMethodNotAllowed:
		errorCode = i18nresp.CodeMethodNotAllowed
	case http.StatusRequestTimeout:
		errorCode = i18nresp.CodeRequestTimeout
	case http.StatusTooManyRequests:
		errorCode = i18nresp.CodeTooManyRequests
	case http.StatusInternalServerError:
		errorCode = i18nresp.CodeInternalError
	case http.StatusNotImplemented:
		errorCode = i18nresp.CodeNotImplemented
	case http.StatusServiceUnavailable:
		errorCode = i18nresp.CodeServiceUnavailable
	case http.StatusGatewayTimeout:
		errorCode = i18nresp.CodeGatewayTimeout
	default:
		errorCode = i18nresp.CodeInternalError
	}

	i18nresp.ErrorResponse(c, errorCode, "")
}

// LocalizedSuccess Return localized success response
func LocalizedSuccess(c *gin.Context, data interface{}) {
	i18nresp.SuccessResponse(c, data)
}

// SetLanguageCookie Set language cookie
func SetLanguageCookie(c *gin.Context, lang string) {
	// Set Cookie, valid for 30 days
	c.SetCookie("lang", lang, 60*60*24*30, "/", "", false, true)
}

// GetLanguageFromRequest Get language code from request
func GetLanguageFromRequest(c *gin.Context) string {
	// Use pkg/i18n language detection logic
	lang := i18nresp.GetLanguageFromGin(c)
	return string(lang)
}
