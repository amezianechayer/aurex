package actions

import (
	"net/http"
	"strconv"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/wallets"
	"github.com/gin-gonic/gin"
)

// WalletController exposes the wallet primitive. It owns no state: each handler
// builds a wallets.Service bound to the request's ledger, so every balance and
// movement is the ledger's truth and passes the Guard at Commit.
type WalletController struct {
	BaseController
}

func NewWalletController() WalletController { return WalletController{} }

func (ctl *WalletController) service(c *gin.Context) *wallets.Service {
	l, _ := c.Get("ledger")
	led := l.(*ledger.Ledger)
	return wallets.NewService(led, led.Store())
}

// responseWalletError renders a typed wallet error, a guard deny, or falls back
// to 500 — mirroring the contract controller's policy-error handling.
func (ctl *WalletController) responseWalletError(c *gin.Context, err error) {
	if we, ok := err.(*wallets.Error); ok {
		out := gin.H{"ok": false, "error": we.Code, "message": we.Message}
		if we.WalletID != "" {
			out["wallet_id"] = we.WalletID
		}
		if we.HoldID != "" {
			out["hold_id"] = we.HoldID
		}
		c.AbortWithStatusJSON(we.HTTPStatus(), out)
		return
	}
	if ge, ok := err.(*guard.Error); ok {
		ctl.responseGuardError(c, ge)
		return
	}
	ctl.responseError(c, http.StatusInternalServerError, err)
}

type createWalletRequest struct {
	Owner string `json:"owner"`
	Asset string `json:"asset"`
}

func (ctl *WalletController) PostWallet(c *gin.Context) {
	var req createWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseWalletError(c, &wallets.Error{Code: wallets.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	w, err := ctl.service(c).Create(req.Owner, req.Asset)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusCreated, w)
}

func (ctl *WalletController) GetWallet(c *gin.Context) {
	w, err := ctl.service(c).Get(c.Param("id"))
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, w)
}

func (ctl *WalletController) ListWallets(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	list, err := ctl.service(c).List(limit, offset)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}

func (ctl *WalletController) GetBalances(c *gin.Context) {
	b, err := ctl.service(c).Balances(c.Param("id"))
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, b)
}

type moveRequest struct {
	Asset       string `json:"asset"`
	Amount      int64  `json:"amount"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

func (ctl *WalletController) PostCredit(c *gin.Context) {
	var req moveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseWalletError(c, &wallets.Error{Code: wallets.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	tx, err := ctl.service(c).Credit(c.Param("id"), req.Asset, req.Amount, req.Source)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, tx)
}

func (ctl *WalletController) PostDebit(c *gin.Context) {
	var req moveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseWalletError(c, &wallets.Error{Code: wallets.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	tx, err := ctl.service(c).Debit(c.Param("id"), req.Asset, req.Amount, req.Destination)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, tx)
}

type holdRequest struct {
	Asset       string `json:"asset"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

func (ctl *WalletController) PostHold(c *gin.Context) {
	var req holdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseWalletError(c, &wallets.Error{Code: wallets.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	h, err := ctl.service(c).Hold(c.Param("id"), req.Asset, req.Amount, req.Description)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusCreated, h)
}

func (ctl *WalletController) ListHolds(c *gin.Context) {
	list, err := ctl.service(c).ListHolds(c.Param("id"))
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}

type captureRequest struct {
	Destination string `json:"destination"`
}

func (ctl *WalletController) PostCaptureHold(c *gin.Context) {
	var req captureRequest
	c.ShouldBindJSON(&req)
	tx, err := ctl.service(c).CaptureHold(c.Param("id"), c.Param("hold_id"), req.Destination)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, tx)
}

func (ctl *WalletController) PostVoidHold(c *gin.Context) {
	tx, err := ctl.service(c).VoidHold(c.Param("id"), c.Param("hold_id"))
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, tx)
}
