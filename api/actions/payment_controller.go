package actions

import (
	"net/http"
	"strconv"

	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/payments"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// PaymentController exposes PSP connectors. The connector registry is built once
// from config; the per-request service binds it to the request's ledger +
// wallets so every payin/payout is a ledger posting through the Guard.
type PaymentController struct {
	BaseController
	reg *payments.Registry
}

func NewPaymentController() PaymentController {
	return PaymentController{reg: buildPaymentRegistry()}
}

// buildPaymentRegistry wires the connectors that are configured (via
// CORREN_PAYMENTS_*). An unconfigured PSP simply isn't registered, so calls to
// it return ERR_UNKNOWN_PSP rather than failing with half-set credentials.
func buildPaymentRegistry() *payments.Registry {
	var conns []payments.PSP
	if viper.GetString("payments.stripe.webhook_secret") != "" || viper.GetString("payments.stripe.secret_key") != "" {
		conns = append(conns, payments.NewStripe(payments.StripeConfig{
			SecretKey:     viper.GetString("payments.stripe.secret_key"),
			WebhookSecret: viper.GetString("payments.stripe.webhook_secret"),
			BaseURL:       viper.GetString("payments.stripe.base_url"),
		}))
	}
	if viper.GetString("payments.paytabs.server_key") != "" {
		conns = append(conns, payments.NewPayTabs(payments.PayTabsConfig{
			ServerKey: viper.GetString("payments.paytabs.server_key"),
			ProfileID: viper.GetString("payments.paytabs.profile_id"),
			BaseURL:   viper.GetString("payments.paytabs.base_url"),
		}))
	}
	return payments.NewRegistry(conns...)
}

func (ctl *PaymentController) service(c *gin.Context) *payments.Service {
	l, _ := c.Get("ledger")
	led := l.(*ledger.Ledger)
	// The ledger itself is the LedgerPort: payins/payouts are committed as
	// two-leg transactions through it (and thus through the Guard).
	return payments.NewService(led, led.Store(), ctl.reg)
}

func (ctl *PaymentController) responsePaymentError(c *gin.Context, err error) {
	if pe, ok := err.(*payments.Error); ok {
		out := gin.H{"ok": false, "error": pe.Code, "message": pe.Message}
		if pe.PSP != "" {
			out["psp"] = pe.PSP
		}
		c.AbortWithStatusJSON(pe.HTTPStatus(), out)
		return
	}
	if ge, ok := err.(*guard.Error); ok {
		ctl.responseGuardError(c, ge)
		return
	}
	ctl.responseError(c, http.StatusInternalServerError, err)
}

// PostWebhook ingests a PSP webhook (POST /payments/webhooks/:psp).
func (ctl *PaymentController) PostWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		ctl.responsePaymentError(c, &payments.Error{Code: payments.ErrInvalidPayload, Message: "cannot read body"})
		return
	}
	sig := c.GetHeader("Stripe-Signature")
	if sig == "" {
		sig = c.GetHeader("signature")
	}
	ev, err := ctl.service(c).HandleWebhook(c.Param("psp"), body, sig)
	if err != nil {
		ctl.responsePaymentError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, ev)
}

type payoutRequest struct {
	PSP         string `json:"psp"`
	WalletID    string `json:"wallet_id"`
	Asset       string `json:"asset"`
	Destination string `json:"destination"`
	Amount      int64  `json:"amount"`
}

// PostPayout initiates an outbound payment (POST /payments/payouts).
func (ctl *PaymentController) PostPayout(c *gin.Context) {
	var req payoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctl.responsePaymentError(c, &payments.Error{Code: payments.ErrInvalidParams, Message: "invalid request body"})
		return
	}
	rec, err := ctl.service(c).CreatePayout(req.PSP, req.WalletID, req.Asset, req.Destination, req.Amount)
	if err != nil {
		ctl.responsePaymentError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, rec)
}

// GetPayments lists payment records (GET /payments).
func (ctl *PaymentController) GetPayments(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 200 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	list, err := ctl.service(c).ListPayments(limit, offset)
	if err != nil {
		ctl.responsePaymentError(c, err)
		return
	}
	ctl.response(c, http.StatusOK, list)
}
