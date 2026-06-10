package actions

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/sharia"
	"github.com/gin-gonic/gin"
)

// ContractController -
type ContractController struct {
	BaseController
	engines *engineCache
}

// engineCache keeps one sharia engine per ledger so the per-contract
// locks survive across requests.
type engineCache struct {
	mu      sync.Mutex
	engines map[string]*sharia.Engine
}

func (c *engineCache) get(l *ledger.Ledger) *sharia.Engine {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.engines[l.Name()]; !ok {
		c.engines[l.Name()] = sharia.NewEngine(l, l.Store())
	}
	return c.engines[l.Name()]
}

// NewContractController -
func NewContractController() ContractController {
	return ContractController{
		engines: &engineCache{engines: map[string]*sharia.Engine{}},
	}
}

func (ctl *ContractController) engine(c *gin.Context) (*sharia.Engine, *ledger.Ledger) {
	l, _ := c.Get("ledger")
	led := l.(*ledger.Ledger)
	return ctl.engines.get(led), led
}

// responseShariaError renders the error format of spec §10.
func (ctl *ContractController) responseShariaError(c *gin.Context, err error) {
	if se, ok := err.(*sharia.Error); ok {
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
		if se.AuditSeq > 0 || se.Code == sharia.ErrShariaViolation {
			out["audit_seq"] = se.AuditSeq
		}
		c.AbortWithStatusJSON(se.HTTPStatus(), out)
		return
	}
	ctl.responseError(c, http.StatusInternalServerError, err)
}

// PostContract godoc
// @Summary Create Sharia Contract
// @Description Create a new Murabaha contract (state PROMISE, no posting)
// @Tags contracts
// @Schemes
// @Param ledger path string true "ledger"
// @Param contract body sharia.CreateRequest true "contract"
// @Accept json
// @Produce json
// @Success 201 {object} actions.BaseResponse
// @Failure 400 {object} actions.BaseResponse
// @Router /{ledger}/contracts [post]
func (ctl *ContractController) PostContract(c *gin.Context) {
	engine, _ := ctl.engine(c)

	var req sharia.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseShariaError(c, &sharia.Error{
			Code:    sharia.ErrInvalidParams,
			Message: "invalid request body: " + err.Error(),
		})
		return
	}

	contract, schedule, err := engine.Create(req)
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}

	ctl.response(c, http.StatusCreated, gin.H{
		"contract": contract,
		"schedule": schedule,
	})
}

// ListContracts godoc
// @Summary List Sharia Contracts
// @Description List sharia contracts
// @Tags contracts
// @Schemes
// @Param ledger path string true "ledger"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/contracts [get]
func (ctl *ContractController) ListContracts(c *gin.Context) {
	_, l := ctl.engine(c)

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	contracts, err := l.Store().ListContracts(limit, offset)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, contracts)
}

// GetContract godoc
// @Summary Get Sharia Contract
// @Description Get a sharia contract with its schedule
// @Tags contracts
// @Schemes
// @Param ledger path string true "ledger"
// @Param id path string true "contract id"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 404 {object} actions.BaseResponse
// @Router /{ledger}/contracts/{id} [get]
func (ctl *ContractController) GetContract(c *gin.Context) {
	engine, _ := ctl.engine(c)

	contract, schedule, err := engine.GetContract(c.Param("id"))
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, gin.H{
		"contract": contract,
		"schedule": schedule,
	})
}

// PostTransition godoc
// @Summary Execute Contract Transition
// @Description Execute a contract FSM transition (acquire, sell, pay_installment, early_settle, late_penalty, cancel)
// @Tags contracts
// @Schemes
// @Param ledger path string true "ledger"
// @Param id path string true "contract id"
// @Param name path string true "transition name"
// @Param input body sharia.TransitionInput false "transition input"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 409 {object} actions.BaseResponse
// @Failure 422 {object} actions.BaseResponse
// @Router /{ledger}/contracts/{id}/transitions/{name} [post]
func (ctl *ContractController) PostTransition(c *gin.Context) {
	engine, _ := ctl.engine(c)

	var input sharia.TransitionInput
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

// GetSchedule godoc
// @Summary Get Contract Schedule
// @Description Get the installment schedule of a contract
// @Tags contracts
// @Schemes
// @Param ledger path string true "ledger"
// @Param id path string true "contract id"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 404 {object} actions.BaseResponse
// @Router /{ledger}/contracts/{id}/schedule [get]
func (ctl *ContractController) GetSchedule(c *gin.Context) {
	engine, _ := ctl.engine(c)

	_, schedule, err := engine.GetContract(c.Param("id"))
	if err != nil {
		ctl.responseShariaError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, schedule)
}

// GetAudit godoc
// @Summary Get Contract Audit Trail
// @Description Get the hash-chained audit trail; pass verify=true to re-verify the chain
// @Tags contracts
// @Schemes
// @Param ledger path string true "ledger"
// @Param id path string true "contract id"
// @Param verify query bool false "re-verify the hash chain"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/contracts/{id}/audit [get]
func (ctl *ContractController) GetAudit(c *gin.Context) {
	engine, _ := ctl.engine(c)

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
