package actions

import (
	"net/http"
	"strconv"

	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
)

// LensController serves the FaRl Lens read-only analytics endpoints.
type LensController struct {
	BaseController
}

// NewLensController -
func NewLensController() LensController { return LensController{} }

func (ctl *LensController) led(c *gin.Context) *ledger.Ledger {
	l, _ := c.Get("ledger")
	return l.(*ledger.Ledger)
}

// Overview godoc
// @Summary Lens Overview
// @Description Aggregate metrics: transaction count, account count, volume by asset, top accounts
// @Tags lens
// @Param ledger path string true "ledger"
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/lens/overview [get]
func (ctl *LensController) Overview(c *gin.Context) {
	ov, err := ctl.led(c).Store().LensOverview()
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, ov)
}

// Flows godoc
// @Summary Lens Flows
// @Description Money-flow graph edges (source→destination per asset per day)
// @Tags lens
// @Param ledger path string true "ledger"
// @Param limit query int false "max edges (1-1000, default 100)"
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/lens/flows [get]
func (ctl *LensController) Flows(c *gin.Context) {
	limit := 100
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			if n < 1 {
				n = 1
			}
			if n > 1000 {
				n = 1000
			}
			limit = n
		}
	}
	edges, err := ctl.led(c).Store().LensFlows(limit)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, edges)
}

// Rollup godoc
// @Summary Lens Rollup
// @Description Contract state counts and guard event action counts
// @Tags lens
// @Param ledger path string true "ledger"
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/lens/rollup [get]
func (ctl *LensController) Rollup(c *gin.Context) {
	r, err := ctl.led(c).Store().LensRollup()
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, r)
}

// TimeSeries godoc
// @Summary Lens TimeSeries
// @Description Daily in/out volume for a given account+asset pair
// @Tags lens
// @Param ledger path string true "ledger"
// @Param account query string true "account address"
// @Param asset query string true "asset code"
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 400 {object} actions.BaseResponse
// @Router /{ledger}/lens/timeseries [get]
func (ctl *LensController) TimeSeries(c *gin.Context) {
	account := c.Query("account")
	asset := c.Query("asset")
	if account == "" || asset == "" {
		ctl.responseError(c, http.StatusBadRequest, errLensParams)
		return
	}
	ts, err := ctl.led(c).Store().LensTimeSeries(account, asset)
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, ts)
}

type lensParamError struct{}

func (lensParamError) Error() string { return "account and asset query params are required" }

var errLensParams = lensParamError{}
