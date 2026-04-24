package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/storage"
	"github.com/gin-gonic/gin"
)

type createKeyRequest struct {
	Name      string           `json:"name" binding:"required"`
	Tier      core.APIKeyTier  `json:"tier"`
	Role      string           `json:"role"`
	ExpiresAt string           `json:"expires_at"`
}

func requireAdmin(c *gin.Context) bool {
	role, _ := c.Get("api_key_role")
	if role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"ok":  false,
			"err": "admin role required",
		})
		return false
	}
	return true
}

func registerAPIKeyRoutes(r *gin.Engine, store storage.Store) {
	// POST /api-keys — create a new key (admin only)
	r.POST("/api-keys", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}

		var req createKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "err": err.Error()})
			return
		}

		if req.Tier == "" {
			req.Tier = core.TierSandbox
		}
		if req.Role == "" {
			req.Role = "reader"
		}

		// Validate tier
		switch req.Tier {
		case core.TierSandbox, core.TierStarter, core.TierGrowth, core.TierEnterprise:
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"ok":  false,
				"err": fmt.Sprintf("invalid tier '%s': must be sandbox|starter|growth|enterprise", req.Tier),
			})
			return
		}

		raw := make([]byte, 24)
		if _, err := rand.Read(raw); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "err": "entropy failure"})
			return
		}
		plain := fmt.Sprintf("crn_%s_%s", req.Tier, hex.EncodeToString(raw))
		hash := HashAPIKey(plain)

		k := core.APIKey{
			KeyHash:   hash,
			Name:      req.Name,
			Role:      req.Role,
			Tier:      req.Tier,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			ExpiresAt: req.ExpiresAt,
		}

		if err := store.SaveAPIKey(k); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "err": err.Error()})
			return
		}

		limit := core.TierRateLimit[req.Tier]
		limitLabel := fmt.Sprintf("%d req/h", limit)
		if limit == 0 {
			limitLabel = "unlimited"
		}

		c.JSON(http.StatusCreated, gin.H{
			"ok":  true,
			"key": plain, // shown once — store it securely
			"meta": gin.H{
				"name":       k.Name,
				"tier":       k.Tier,
				"role":       k.Role,
				"rate_limit": limitLabel,
				"created_at": k.CreatedAt,
				"expires_at": k.ExpiresAt,
			},
		})
	})

	// GET /api-keys — list all keys (admin only, key_hash hidden)
	r.GET("/api-keys", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}

		keys, err := store.ListAPIKeys()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "err": err.Error()})
			return
		}
		if keys == nil {
			keys = []core.APIKey{}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "keys": keys})
	})

	// DELETE /api-keys/:hash — revoke a key by its SHA256 hash (admin only)
	r.DELETE("/api-keys/:hash", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}

		if err := store.DeleteAPIKey(c.Param("hash")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}
