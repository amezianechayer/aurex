package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/storage"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// HashAPIKey returns the SHA256 hex of a plain API key.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// slidingWindow tracks per-key request timestamps for rate limiting.
type slidingWindow struct {
	mu      sync.Mutex
	windows map[string][]time.Time
}

var rateWindows = &slidingWindow{windows: make(map[string][]time.Time)}

// check records a request and returns (remaining, resetUnix).
// remaining < 0 means rate limit exceeded.
func (sw *slidingWindow) check(keyHash string, limit int) (int, int64) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Hour)

	prev := sw.windows[keyHash]
	fresh := prev[:0]
	for _, t := range prev {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	fresh = append(fresh, now)
	sw.windows[keyHash] = fresh

	remaining := limit - len(fresh)
	return remaining, now.Add(time.Hour).Unix()
}

// APIKeyMiddleware authenticates every request via X-API-Key.
// Public paths (/_healthcheck, /_info) are always allowed.
// If no auth is configured at all, runs in open mode.
func APIKeyMiddleware(store storage.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/_healthcheck" || path == "/_info" {
			c.Next()
			return
		}

		adminKey := viper.GetString("server.http.admin_key")
		legacyKeys := viper.GetStringMapString("server.http.api_keys")
		authRequired := adminKey != "" || len(legacyKeys) > 0

		key := c.GetHeader("X-API-Key")
		if key == "" {
			if !authRequired {
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"ok":  false,
				"err": "missing X-API-Key header",
			})
			return
		}

		// Bootstrap admin key from config
		if adminKey != "" && key == adminKey {
			c.Set("api_key_role", "admin")
			c.Set("api_key_tier", core.TierEnterprise)
			c.Next()
			return
		}

		// Legacy static keys from config (backwards compat)
		if role, ok := legacyKeys[key]; ok {
			c.Set("api_key_role", role)
			c.Set("api_key_tier", core.TierEnterprise)
			c.Next()
			return
		}

		// DB lookup by SHA256 hash
		hash := HashAPIKey(key)
		k, err := store.FindAPIKey(hash)
		if err != nil || k == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"ok":  false,
				"err": "invalid API key",
			})
			return
		}

		// Expiry check
		if k.ExpiresAt != "" {
			exp, _ := time.Parse(time.RFC3339, k.ExpiresAt)
			if time.Now().After(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"ok":  false,
					"err": "API key expired",
				})
				return
			}
		}

		// Rate limiting
		limit := core.TierRateLimit[k.Tier]
		if limit > 0 {
			remaining, reset := rateWindows.check(hash, limit)
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, remaining)))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
			if remaining < 0 {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"ok":   false,
					"err":  "rate limit exceeded — upgrade your tier",
					"tier": k.Tier,
				})
				return
			}
		}

		c.Set("api_key", k)
		c.Set("api_key_role", k.Role)
		c.Set("api_key_tier", k.Tier)
		c.Next()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
