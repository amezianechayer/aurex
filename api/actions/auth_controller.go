package actions

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/amezianechayer/corren/auth"
	"github.com/gin-gonic/gin"
)

var errMissingCredentials = errors.New("username and password are required")

// AuthController -
type AuthController struct {
	BaseController
	service *auth.Service
}

// NewAuthController -
func NewAuthController(service *auth.Service) AuthController {
	return AuthController{service: service}
}

func (ctl *AuthController) bearerIdentity(c *gin.Context) (auth.Identity, string, bool) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return auth.Identity{}, "", false
	}
	token := strings.TrimPrefix(header, "Bearer ")
	id, err := ctl.service.Authenticate(token)
	if err != nil {
		return auth.Identity{}, "", false
	}
	return id, token, true
}

// Login godoc
// @Summary Login
// @Description Authenticate a user, returns a revocable session token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 401 {object} actions.BaseResponse
// @Router /auth/login [post]
func (ctl *AuthController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		ctl.responseError(c, http.StatusBadRequest, errMissingCredentials)
		return
	}

	token, expiresAt, err := ctl.service.Login(req.Username, req.Password)
	if err != nil {
		ctl.responseError(c, http.StatusUnauthorized, err)
		return
	}

	id, _ := ctl.service.Authenticate(token)
	ctl.response(c, http.StatusOK, gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"role":       id.Role,
	})
}

// Logout godoc
// @Summary Logout
// @Description Revoke the current session token
// @Tags auth
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /auth/logout [post]
func (ctl *AuthController) Logout(c *gin.Context) {
	_, token, ok := ctl.bearerIdentity(c)
	if !ok {
		ctl.responseError(c, http.StatusUnauthorized, auth.ErrUnauthorized)
		return
	}
	if err := ctl.service.Logout(token); err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, gin.H{"logged_out": true})
}

// Me godoc
// @Summary Current identity
// @Description Return the identity behind the bearer token (used by Horizon)
// @Tags auth
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 401 {object} actions.BaseResponse
// @Router /auth/me [get]
func (ctl *AuthController) Me(c *gin.Context) {
	id, _, ok := ctl.bearerIdentity(c)
	if !ok {
		ctl.responseError(c, http.StatusUnauthorized, auth.ErrUnauthorized)
		return
	}
	ctl.response(c, http.StatusOK, id)
}

// CreateKey godoc
// @Summary Create API key
// @Description Create an API key; the plaintext key is returned ONCE
// @Tags auth
// @Accept json
// @Produce json
// @Success 201 {object} actions.BaseResponse
// @Router /auth/admin/keys [post]
func (ctl *AuthController) CreateKey(c *gin.Context) {
	var req struct {
		Label   string   `json:"label"`
		Role    string   `json:"role"`
		Ledgers []string `json:"ledgers"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseError(c, http.StatusBadRequest, err)
		return
	}

	plain, key, err := ctl.service.CreateKey(req.Label, req.Role, req.Ledgers)
	if err != nil {
		ctl.responseError(c, http.StatusBadRequest, err)
		return
	}

	ctl.response(c, http.StatusCreated, gin.H{
		"id":      key.ID,
		"key":     plain, // shown once, stored hashed
		"label":   key.Label,
		"role":    key.Role,
		"ledgers": key.Ledgers,
	})
}

// ListKeys godoc
// @Summary List API keys
// @Tags auth
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /auth/admin/keys [get]
func (ctl *AuthController) ListKeys(c *gin.Context) {
	keys, err := ctl.service.ListKeys()
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, keys)
}

// RevokeKey godoc
// @Summary Revoke API key
// @Tags auth
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /auth/admin/keys/{id} [delete]
func (ctl *AuthController) RevokeKey(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ctl.responseError(c, http.StatusBadRequest, err)
		return
	}
	if err := ctl.service.RevokeKey(id); err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, gin.H{"revoked": true})
}

// CreateUser godoc
// @Summary Create user
// @Tags auth
// @Accept json
// @Produce json
// @Success 201 {object} actions.BaseResponse
// @Router /auth/admin/users [post]
func (ctl *AuthController) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		ctl.responseError(c, http.StatusBadRequest, errMissingCredentials)
		return
	}

	user, err := ctl.service.CreateUser(req.Username, req.Password, req.Role)
	if err != nil {
		ctl.responseError(c, http.StatusBadRequest, err)
		return
	}
	ctl.response(c, http.StatusCreated, user)
}

// ListUsers godoc
// @Summary List users
// @Tags auth
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /auth/admin/users [get]
func (ctl *AuthController) ListUsers(c *gin.Context) {
	users, err := ctl.service.ListUsers()
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, users)
}
