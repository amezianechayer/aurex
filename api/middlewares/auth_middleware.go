package middlewares

import (
	"net/http"
	"strings"

	"github.com/amezianechayer/corren/auth"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// AuthMiddleware struct
type AuthMiddleware struct {
	service *auth.Service
}

// NewAuthMiddleware
func NewAuthMiddleware(service *auth.Service) AuthMiddleware {
	return AuthMiddleware{service: service}
}

// AuthMiddleware enforces bearer authentication, the role matrix
// (readonly for reads, operator for writes, admin for /auth/admin) and
// the per-ledger key scope. Disabled by default (auth.enabled=false):
// in that mode behavior is identical to the historical server.
func (m AuthMiddleware) AuthMiddleware(engine *gin.Engine) gin.HandlerFunc {
	// legacy option, kept for backward compatibility
	if basic := viper.Get("server.http.basic_auth"); basic != nil {
		segment := strings.Split(basic.(string), ":")
		if len(segment) == 2 {
			engine.Use(gin.BasicAuth(gin.Accounts{segment[0]: segment[1]}))
		}
	}

	return func(c *gin.Context) {
		if !viper.GetBool("auth.enabled") {
			return
		}

		path := c.FullPath()
		if path == "/auth/login" || path == "/healthz" || path == "/docs" || path == "/openapi.yaml" {
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			abortAuth(c, http.StatusUnauthorized, "missing bearer token")
			return
		}

		identity, err := m.service.Authenticate(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			abortAuth(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		need := auth.RoleReadonly
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			need = auth.RoleOperator
		}
		if strings.HasPrefix(path, "/auth/admin") {
			need = auth.RoleAdmin
		}
		if !auth.RoleAtLeast(identity.Role, need) {
			abortAuth(c, http.StatusForbidden, "role "+identity.Role+" cannot access this resource")
			return
		}

		if ledger := c.Param("ledger"); ledger != "" && !identity.AllowsLedger(ledger) {
			abortAuth(c, http.StatusForbidden, "credentials not scoped to ledger "+ledger)
			return
		}

		c.Set("identity", identity)
	}
}

func abortAuth(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"ok":            false,
		"error":         true,
		"error_code":    status,
		"error_message": msg,
	})
}
