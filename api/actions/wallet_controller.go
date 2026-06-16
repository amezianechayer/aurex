package actions

import (
	"net/http"
	"strconv"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/ledger/query"
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

func (ctl *WalletController) ledgerOf(c *gin.Context) *ledger.Ledger {
	l, _ := c.Get("ledger")
	return l.(*ledger.Ledger)
}

func (ctl *WalletController) service(c *gin.Context) *wallets.Service {
	led := ctl.ledgerOf(c)
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

// GetWallet returns the wallet together with its current ledger balance.
func (ctl *WalletController) GetWallet(c *gin.Context) {
	svc := ctl.service(c)
	w, err := svc.Get(c.Param("id"))
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	b, err := svc.Balances(w.ID)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, gin.H{
		"id":         w.ID,
		"owner":      w.Owner,
		"asset":      w.Asset,
		"created_at": w.CreatedAt,
		"updated_at": w.UpdatedAt,
		"balances":   gin.H{"available": b.Available, "held": b.Held},
	})
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

type transferRequest struct {
	To     string `json:"to"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
}

// PostTransfer moves funds from this wallet to another wallet.
func (ctl *WalletController) PostTransfer(c *gin.Context) {
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseWalletError(c, &wallets.Error{Code: wallets.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	tx, err := ctl.service(c).Transfer(c.Param("id"), req.To, req.Asset, req.Amount)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, tx)
}

// GetTransactions returns the ledger transactions touching this wallet's main
// account (topups, debits, transfers, hold movements).
func (ctl *WalletController) GetTransactions(c *gin.Context) {
	svc := ctl.service(c)
	if _, err := svc.Get(c.Param("id")); err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	cursor, err := ctl.ledgerOf(c).FindTransactions(query.Account(wallets.MainAccount(c.Param("id"))))
	if err != nil {
		ctl.responseError(c, http.StatusInternalServerError, err)
		return
	}
	ctl.response(c, http.StatusOK, cursor)
}

type holdRequest struct {
	Asset     string `json:"asset"`
	Amount    int64  `json:"amount"`
	Reason    string `json:"reason"`
	ExpiresAt string `json:"expires_at"`
}

func (ctl *WalletController) PostHold(c *gin.Context) {
	var req holdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responseWalletError(c, &wallets.Error{Code: wallets.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	h, err := ctl.service(c).Hold(c.Param("id"), req.Asset, req.Amount, req.Reason, req.ExpiresAt)
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusCreated, h)
}

// ReleaseHold frees a hold back to the wallet (DELETE /holds/:hold_id).
func (ctl *WalletController) ReleaseHold(c *gin.Context) {
	tx, err := ctl.service(c).Release(c.Param("id"), c.Param("hold_id"))
	if err != nil {
		ctl.responseWalletError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, tx)
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
