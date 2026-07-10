package actions

import (
	"strings"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
)

// ScriptController -
type ScriptController struct {
	BaseController
}

// NewScriptController -
func NewScriptController() ScriptController {
	return ScriptController{}
}

// PostScript godoc
// @Summary Execute Farl
// @Description Execute a Farl and create the transaction if any
// @Tags script
// @Schemes
// @Param ledger path string true "ledger"
// @Param script body core.Script true "script"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/script [post]
func (ctl *ScriptController) PostScript(c *gin.Context) {
	l, _ := c.Get("ledger")

	var script core.Script
	c.ShouldBind(&script)

	err := l.(*ledger.Ledger).Execute(c, script)

	res := gin.H{
		"ok": err == nil,
	}

	if err != nil {
		err_str := err.Error()
		err_str = strings.ReplaceAll(err_str, "\n", "\r\n")
		res["err"] = err_str
	}

	c.JSON(200, res)
}
