package actions

import (
	"encoding/json"
	"net/http"

	"github.com/amezianechayer/corren/flows"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/shariawire"
	"github.com/gin-gonic/gin"
)

// FlowController exposes the orchestration engine. Each handler builds a
// flows.Engine bound to the request's ledger and the corren-sharia engine
// (resolved via shariawire), so steps run against the right ledger (through the
// Guard) and the right contract engine.
type FlowController struct {
	BaseController
}

func NewFlowController() FlowController { return FlowController{} }

func (ctl *FlowController) engine(c *gin.Context) (*flows.Engine, error) {
	l, _ := c.Get("ledger")
	led := l.(*ledger.Ledger)
	she, err := shariawire.EngineFor(led)
	if err != nil {
		return nil, err
	}
	return flows.NewEngine(led, she, led.Store()), nil
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
	eng, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	var req createFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseFlowError(c, &flows.Error{Code: flows.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	f, err := eng.CreateFlow(req.Name, req.Trigger, req.Schedule, req.Steps)
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusCreated, f)
}

func (ctl *FlowController) ListFlows(c *gin.Context) {
	eng, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	list, err := eng.ListFlows()
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}

func (ctl *FlowController) GetFlow(c *gin.Context) {
	eng, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	f, err := eng.GetFlow(c.Param("id"))
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, f)
}

// PostTrigger starts a manual run; the whole request body is the flow input.
func (ctl *FlowController) PostTrigger(c *gin.Context) {
	eng, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	body, _ := c.GetRawData()
	var input json.RawMessage
	if len(body) > 0 {
		input = json.RawMessage(body)
	}
	inst, err := eng.Trigger(c.Param("id"), input)
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
	eng, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	if _, err := eng.GetFlow(c.Param("id")); err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	list, err := eng.ListInstances(c.Param("id"))
	if err != nil {
		ctl.responseFlowError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}
