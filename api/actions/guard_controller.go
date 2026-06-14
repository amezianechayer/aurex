package actions

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
)

type GuardController struct {
	BaseController
}

func NewGuardController() GuardController { return GuardController{} }

func (ctl *GuardController) led(c *gin.Context) *ledger.Ledger {
	l, _ := c.Get("ledger")
	return l.(*ledger.Ledger)
}

func (ctl *GuardController) responseGuardError(c *gin.Context, err error) {
	if ge, ok := err.(*guard.Error); ok {
		out := gin.H{"error": ge.Code, "message": ge.Message}
		if ge.StandardRef != "" {
			out["standard_ref"] = ge.StandardRef
		}
		if ge.RuleID != "" {
			out["rule_id"] = ge.RuleID
		}
		c.AbortWithStatusJSON(ge.HTTPStatus(), out)
		return
	}
	ctl.responseError(c, http.StatusInternalServerError, err)
}

func (ctl *GuardController) PostRule(c *gin.Context) {
	l := ctl.led(c)
	var r guard.Rule
	if err := c.ShouldBindJSON(&r); err != nil {
		ctl.responseGuardError(c, &guard.Error{Code: guard.ErrInvalidRule, Message: "invalid rule body: " + err.Error()})
		return
	}
	if r.ID == "" {
		r.ID = fmt.Sprintf("rule-%d", time.Now().UnixNano())
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	r.CreatedAt, r.UpdatedAt = ts, ts
	if err := guard.ValidateRule(&r); err != nil {
		ctl.responseGuardError(c, err)
		return
	}
	if err := l.Store().SaveRule(r); err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	l.Guard().Reload()
	ctl.response(c, http.StatusCreated, r)
}

func (ctl *GuardController) ListRules(c *gin.Context) {
	rules, err := ctl.led(c).Store().ListRules()
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, rules)
}

func (ctl *GuardController) GetRule(c *gin.Context) {
	r, err := ctl.led(c).Store().GetRule(c.Param("id"))
	if err != nil {
		ctl.responseGuardError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, r)
}

func (ctl *GuardController) PatchRule(c *gin.Context) {
	l := ctl.led(c)
	existing, err := l.Store().GetRule(c.Param("id"))
	if err != nil {
		ctl.responseGuardError(c, err)
		return
	}
	// bind onto the existing rule (only provided fields change)
	if err := c.ShouldBindJSON(&existing); err != nil {
		ctl.responseGuardError(c, &guard.Error{Code: guard.ErrInvalidRule, Message: "invalid rule body: " + err.Error()})
		return
	}
	existing.ID = c.Param("id")
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := guard.ValidateRule(&existing); err != nil {
		ctl.responseGuardError(c, err)
		return
	}
	if err := l.Store().UpdateRule(existing); err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	l.Guard().Reload()
	ctl.response(c, http.StatusOK, existing)
}

func (ctl *GuardController) DeleteRule(c *gin.Context) {
	l := ctl.led(c)
	if err := l.Store().DeleteRule(c.Param("id")); err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	l.Guard().Reload()
	ctl.response(c, http.StatusOK, gin.H{"deleted": true})
}

func (ctl *GuardController) ListEvents(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 500 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	events, err := ctl.led(c).Store().ListGuardEvents(limit, offset)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, events)
}
