package actions

import (
	"reflect"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/gin-gonic/gin"
)

// actions struct
type BaseController struct{}
type BaseResponse struct {
	Ok bool `json:"ok"`
}

func (ctl *BaseController) response(c *gin.Context, status int, data interface{}) {
	if data == nil {
		c.Status(status)
	}
	if reflect.TypeOf(data) == reflect.TypeOf(query.Cursor{}) {
		c.JSON(status, gin.H{
			"ok":     true,
			"cursor": data,
		})
	} else {
		c.JSON(status, gin.H{
			"ok":   true,
			"data": data,
		})
	}
}

func (ctl *BaseController) responseError(c *gin.Context, status int, err error) {
	c.Abort()
	c.AbortWithStatusJSON(status, gin.H{
		"ok":            false,
		"error":         true,
		"error_code":    status,
		"error_message": err.Error(),
	})
}

// responseGuardError renders a guard rule rejection. A guard deny is an expected
// policy outcome (like a sharia violation), so it carries the rule's code and
// standard_ref and uses the guard error's own HTTP status (422 for a deny) — not
// a 500, which would look like a server fault to clients and monitoring.
func (ctl *BaseController) responseGuardError(c *gin.Context, ge *guard.Error) {
	out := gin.H{"ok": false, "error": ge.Code, "message": ge.Message}
	if ge.StandardRef != "" {
		out["standard_ref"] = ge.StandardRef
	}
	if ge.RuleID != "" {
		out["rule_id"] = ge.RuleID
	}
	c.AbortWithStatusJSON(ge.HTTPStatus(), out)
}
