package actions

import (
	"net/http"
	"strconv"

	csharia "github.com/amezianechayer/corren-sharia"
	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/shariawire"
	"github.com/gin-gonic/gin"
)

// ContractController is thin HTTP wiring over the standalone corren-sharia
// engine. All Sharia logic lives in corren-sharia; this controller resolves the
// per-ledger engine (via shariawire) and renders responses.
type ContractController struct {
	BaseController
}

// NewContractController -
func NewContractController() ContractController { return ContractController{} }

func (ctl *ContractController) engine(c *gin.Context) (*csharia.Engine, error) {
	l, _ := c.Get("ledger")
	return shariawire.EngineFor(l.(*ledger.Ledger))
}

// responseShariaError renders the error format of spec §10.
func (ctl *ContractController) responseShariaError(c *gin.Context, err error) {
	if se, ok := err.(*csharia.Error); ok {
		out := gin.H{
			"error":   se.Code,
			"message": se.Message,
		}
		if se.StandardRef != "" {
			out["standard_ref"] = se.StandardRef
		}
		if se.ContractID != "" {
			out["contract_id"] = se.ContractID
		}
		if se.AuditSeq > 0 || se.Code == csharia.ErrShariaViolation {
			out["audit_seq"] = se.AuditSeq
		}
		c.AbortWithStatusJSON(se.HTTPStatus(), out)
		return
	}
	if ge, ok := err.(*guard.Error); ok {
		// A FaRl Guard rule denied the transition's postings at ledger.Commit.
		ctl.responseGuardError(c, ge)
		return
	}
	ctl.responseError(c, http.StatusInternalServerError, err)
}

func (ctl *ContractController) PostContract(c *gin.Context) {
	engine, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	var req csharia.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseShariaError(c, &csharia.Error{Code: csharia.ErrInvalidParams, Message: "invalid request body: " + err.Error()})
		return
	}
	contract, schedule, err := engine.Create(req)
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	ctl.response(c, http.StatusCreated, gin.H{"contract": contract, "schedule": schedule})
}

func (ctl *ContractController) ListContracts(c *gin.Context) {
	engine, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	limit, e := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if e != nil || limit < 1 || limit > 200 {
		limit = 50
	}
	offset, e := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if e != nil || offset < 0 {
		offset = 0
	}
	contracts, err := engine.ListContracts(limit, offset)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, contracts)
}

func (ctl *ContractController) GetContract(c *gin.Context) {
	engine, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	contract, schedule, err := engine.GetContract(c.Param("id"))
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, gin.H{"contract": contract, "schedule": schedule})
}

func (ctl *ContractController) PostTransition(c *gin.Context) {
	engine, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	var input csharia.TransitionInput
	c.ShouldBindJSON(&input)
	result, err := engine.Transition(c.Param("id"), c.Param("name"), input)
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, gin.H{
		"contract_id": result.Contract.ID,
		"transition":  result.Transition,
		"new_state":   result.NewState,
		"tx_id":       result.TxID,
	})
}

func (ctl *ContractController) GetSchedule(c *gin.Context) {
	engine, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	_, schedule, err := engine.GetContract(c.Param("id"))
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, schedule)
}

func (ctl *ContractController) GetAudit(c *gin.Context) {
	engine, err := ctl.engine(c)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	id := c.Param("id")
	if _, _, err := engine.GetContract(id); err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	events, valid, err := engine.VerifyAudit(id)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	out := gin.H{"events": events}
	if c.Query("verify") == "true" {
		out["chain_valid"] = valid
	}
	ctl.response(c, http.StatusOK, out)
}
