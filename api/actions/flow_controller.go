package actions

import (
	"encoding/json"
	"net/http"

	"github.com/amezianechayer/corren/flows"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
	"github.com/gin-gonic/gin"
)

// FlowController exposes the orchestration engine. Each handler builds a
// flows.Engine bound to the request's ledger + its sharia engine, so steps run
// against the right ledger (and through the Guard).
type FlowController struct {
	BaseController
}

func NewFlowController() FlowController { return FlowController{} }

func (ctl *FlowController) engine(c *gin.Context) *flows.Engine {
	l, _ := c.Get("ledger")
	led := l.(*ledger.Ledger)
	return flows.NewEngine(led, sharia.NewEngine(led, led.Store()), led.Store())
}

func (ctl *FlowController) responseFlowError(c *gin.Context, err error) {
	if fe, ok := err.(*flows.Error); ok {
		out := gin.H{"ok": false, "error": fe.Code, "message": fe.Message}
		if fe.FlowID != "" {
			out["flow_id"] = fe.FlowID
		}
		c.AbortWithStatusJSON(fe.HTTPStatus(), out)
		return
	}
	ctl.responseError(c, http.StatusInternalServerError, err)
}

type createFlowRequest struct {
	Name     string          `json:"name"`
	Trigger  string          `json:"trigger"`
	Schedule string          `json:"schedule"`
	Steps    json.RawMessage `json:"steps"`
}

func (ctl *FlowController) PostFlow(c *gin.Context) {
	var req createFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseFlowError(c, &flows.Error{Code: flows.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	f, err := ctl.engine(c).CreateFlow(req.Name, req.Trigger, req.Schedule, req.Steps)
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusCreated, f)
}

func (ctl *FlowController) ListFlows(c *gin.Context) {
	list, err := ctl.engine(c).ListFlows()
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}

func (ctl *FlowController) GetFlow(c *gin.Context) {
	f, err := ctl.engine(c).GetFlow(c.Param("id"))
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, f)
}

// PostTrigger starts a manual run; the whole request body is the flow input.
func (ctl *FlowController) PostTrigger(c *gin.Context) {
	body, _ := c.GetRawData()
	var input json.RawMessage
	if len(body) > 0 {
		input = json.RawMessage(body)
	}
	inst, err := ctl.engine(c).Trigger(c.Param("id"), input)
	if err != nil {
		if fe, ok := err.(*flows.Error); ok && fe.Code == flows.ErrNotFound {
			ctl.responseFlowError(c, err)
			return
		}
		// A run that failed/stopped still produced an instance — return it with
		// its terminal status rather than a bare error.
		ctl.response(c, http.StatusOK, inst)
		return
	}
	ctl.response(c, http.StatusOK, inst)
}

func (ctl *FlowController) GetInstances(c *gin.Context) {
	if _, err := ctl.engine(c).GetFlow(c.Param("id")); err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	list, err := ctl.engine(c).ListInstances(c.Param("id"))
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}
