package actions

import (
	"errors"
	"net/http"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/gin-gonic/gin"
)

// TransactionController -
type TransactionController struct {
	BaseController
}

// NewTransactionController -
func NewTransactionController() TransactionController {
	return TransactionController{}
}

// GetTransactions godoc
// @Summary Get Transactions
// @Description Get all ledger transactions
// @Tags transactions
// @Schemes
// @Description List transactions
// @Param ledger path string true "ledger"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse{cursor=query.Cursor{data=[]core.Transaction}}
// @Router /{ledger}/transactions [get]
func (ctl *TransactionController) GetTransactions(c *gin.Context) {
	l, _ := c.Get("ledger")
	cursor, err := l.(*ledger.Ledger).FindTransactions(
		query.After(c.Query("after")),
		query.Reference(c.Query("reference")),
		query.Account(c.Query("account")),
	)
	if err != nil {
		ctl.responseError(
			c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	ctl.response(
		c,
		http.StatusOK,
		cursor,
	)
}

// PostTransactions godoc
// @Summary Create Transaction
// @Description Create a new ledger transaction
// @Tags transactions
// @Schemes
// @Description Commit a new transaction to the ledger
// @Param ledger path string true "ledger"
// @Param transaction body core.Transaction true "transaction"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/transactions [post]
func (ctl *TransactionController) PostTransaction(c *gin.Context) {
	l, _ := c.Get("ledger")

	var t core.Transaction
	c.ShouldBind(&t)

	ts, err := l.(*ledger.Ledger).Commit([]core.Transaction{t})
	if err != nil {
		if ge, ok := err.(*guard.Error); ok {
			// A FaRl Guard rule denied this transaction at ledger.Commit.
			ctl.responseGuardError(c, ge)
			return
		}
		ctl.responseError(
			c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	ctl.response(
		c,
		http.StatusOK,
		ts,
	)
}

// GetTransaction godoc
// @Summary Revert Transaction
// @Description Get transaction by transaction id
// @Tags transactions
// @Schemes
// @Param ledger path string true "ledger"
// @Param txid path string true "txid"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Failure 404 {object} actions.BaseResponse
// @Router /{ledger}/transactions/{txid} [get]
func (ctl *TransactionController) GetTransaction(c *gin.Context) {
	l, _ := c.Get("ledger")
	tx, err := l.(*ledger.Ledger).GetTransaction(c.Param("txid"))
	if err != nil {
		ctl.responseError(
			c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	if tx.Postings == nil {
		ctl.responseError(
			c,
			http.StatusNotFound,
			errors.New("transaction not found"),
		)
		return
	}
	ctl.response(
		c,
		http.StatusOK,
		tx,
	)
}

// RevertTransaction godoc
// @Summary Revert Transaction
// @Description Revert a ledger transaction by transaction id
// @Tags transactions
// @Schemes
// @Param ledger path string true "ledger"
// @Param txid path string true "txid"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/transactions/{txid}/revert [post]
func (ctl *TransactionController) RevertTransaction(c *gin.Context) {
	l, _ := c.Get("ledger")
	err := l.(*ledger.Ledger).RevertTransaction(c.Param("txid"))
	if err != nil {
		ctl.responseError(
			c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	ctl.response(
		c,
		http.StatusOK,
		nil,
	)
}

// PostTransactionMetadata godoc
// @Summary Set Transaction Metadata
// @Description Set a new metadata to a ledger transaction by transaction id
// @Tags transactions
// @Schemes
// @Param ledger path string true "ledger"
// @Param txid path string true "txid"
// @Accept json
// @Produce json
// @Success 200 {object} actions.BaseResponse
// @Router /{ledger}/transactions/{txid}/metadata [post]
func (ctl *TransactionController) PostTransactionMetadata(c *gin.Context) {
	l, _ := c.Get("ledger")

	var m core.Metadata
	c.ShouldBind(&m)

	err := l.(*ledger.Ledger).SaveMeta(
		"transaction",
		c.Param("txid"),
		m,
	)
	if err != nil {
		ctl.responseError(
			c,
			http.StatusInternalServerError,
			err,
		)
		return
	}
	ctl.response(
		c,
		http.StatusOK,
		nil,
	)
}
