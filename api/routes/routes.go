package routes

import (
	"github.com/amezianechayer/corren/api/actions"
	"github.com/amezianechayer/corren/api/middlewares"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewRoutes),
)

// Routes -
type Routes struct {
	resolver              *ledger.Resolver
	authMiddleware        middlewares.AuthMiddleware
	ledgerMiddleware      middlewares.LedgerMiddleware
	configController      actions.ConfigController
	ledgerController      actions.LedgerController
	scriptController      actions.ScriptController
	accountController     actions.AccountController
	transactionController actions.TransactionController
	contractController    actions.ContractController
	guardController       actions.GuardController
	authController        actions.AuthController
	lensController        actions.LensController
	walletController      actions.WalletController
	paymentController     actions.PaymentController
	flowController        actions.FlowController
}

// NewRoutes -
func NewRoutes(
	resolver *ledger.Resolver,
	authMiddleware middlewares.AuthMiddleware,
	ledgerMiddleware middlewares.LedgerMiddleware,
	configController actions.ConfigController,
	ledgerController actions.LedgerController,
	scriptController actions.ScriptController,
	accountController actions.AccountController,
	transactionController actions.TransactionController,
	contractController actions.ContractController,
	guardController actions.GuardController,
	authController actions.AuthController,
	lensController actions.LensController,
	walletController actions.WalletController,
	paymentController actions.PaymentController,
	flowController actions.FlowController,
) *Routes {
	return &Routes{
		resolver:              resolver,
		authMiddleware:        authMiddleware,
		ledgerMiddleware:      ledgerMiddleware,
		configController:      configController,
		ledgerController:      ledgerController,
		scriptController:      scriptController,
		accountController:     accountController,
		transactionController: transactionController,
		contractController:    contractController,
		guardController:       guardController,
		authController:        authController,
		lensController:        lensController,
		walletController:      walletController,
		paymentController:     paymentController,
		flowController:        flowController,
	}
}

// Engine -
func (r *Routes) Engine(cc cors.Config) *gin.Engine {
	engine := gin.New()

	// Default Middlewares
	engine.Use(
		cors.New(cc),
		gin.Recovery(),
		logger.SetLogger(),
		r.authMiddleware.AuthMiddleware(engine),
	)

	engine.GET("/swagger.json", r.configController.GetDocs)

	// AuthController
	authGroup := engine.Group("/auth")
	{
		authGroup.POST("/login", r.authController.Login)
		authGroup.POST("/logout", r.authController.Logout)
		authGroup.GET("/me", r.authController.Me)
		authGroup.POST("/admin/keys", r.authController.CreateKey)
		authGroup.GET("/admin/keys", r.authController.ListKeys)
		authGroup.DELETE("/admin/keys/:id", r.authController.RevokeKey)
		authGroup.POST("/admin/users", r.authController.CreateUser)
		authGroup.GET("/admin/users", r.authController.ListUsers)
	}

	// API Routes
	engine.GET("/_info", r.configController.GetInfo)

	ledger := engine.Group("/:ledger", r.ledgerMiddleware.LedgerMiddleware())
	{
		// LedgerController
		ledger.GET("/stats", r.ledgerController.GetStats)

		// TransactionController
		ledger.GET("/transactions", r.transactionController.GetTransactions)
		ledger.POST("/transactions", r.transactionController.PostTransaction)
		ledger.GET("/transactions/:txid", r.transactionController.GetTransaction)
		ledger.POST("/transactions/:txid/revert", r.transactionController.RevertTransaction)
		ledger.POST("/transactions/:txid/metadata", r.transactionController.PostTransactionMetadata)

		// AccountController
		ledger.GET("/accounts", r.accountController.GetAccounts)
		ledger.GET("/accounts/:address", r.accountController.GetAccount)
		ledger.POST("/accounts/:address/metadata", r.accountController.PostAccountMetadata)

		// ScriptController
		ledger.POST("/script", r.scriptController.PostScript)

		// ContractController (sharia)
		ledger.POST("/contracts", r.contractController.PostContract)
		ledger.GET("/contracts", r.contractController.ListContracts)
		ledger.GET("/contracts/:id", r.contractController.GetContract)
		ledger.POST("/contracts/:id/transitions/:name", r.contractController.PostTransition)
		ledger.GET("/contracts/:id/schedule", r.contractController.GetSchedule)
		ledger.GET("/contracts/:id/audit", r.contractController.GetAudit)

		// GuardController
		ledger.POST("/guard/rules", r.guardController.PostRule)
		ledger.GET("/guard/rules", r.guardController.ListRules)
		ledger.GET("/guard/rules/:id", r.guardController.GetRule)
		ledger.PATCH("/guard/rules/:id", r.guardController.PatchRule)
		ledger.DELETE("/guard/rules/:id", r.guardController.DeleteRule)
		ledger.GET("/guard/events", r.guardController.ListEvents)

		// LensController (FaRl Lens read-only analytics)
		ledger.GET("/lens/overview", r.lensController.Overview)
		ledger.GET("/lens/flows", r.lensController.Flows)
		ledger.GET("/lens/rollup", r.lensController.Rollup)
		ledger.GET("/lens/timeseries", r.lensController.TimeSeries)

		// WalletController (user wallets on top of the ledger; Sharia-aware holds)
		ledger.POST("/wallets", r.walletController.PostWallet)
		ledger.GET("/wallets", r.walletController.ListWallets)
		ledger.GET("/wallets/:id", r.walletController.GetWallet)
		ledger.GET("/wallets/:id/balances", r.walletController.GetBalances)
		ledger.GET("/wallets/:id/transactions", r.walletController.GetTransactions)
		ledger.POST("/wallets/:id/credit", r.walletController.PostCredit)
		ledger.POST("/wallets/:id/debit", r.walletController.PostDebit)
		ledger.POST("/wallets/:id/transfer", r.walletController.PostTransfer)
		ledger.POST("/wallets/:id/holds", r.walletController.PostHold)
		ledger.GET("/wallets/:id/holds", r.walletController.ListHolds)
		ledger.DELETE("/wallets/:id/holds/:hold_id", r.walletController.ReleaseHold)
		ledger.POST("/wallets/:id/holds/:hold_id/capture", r.walletController.PostCaptureHold)
		ledger.POST("/wallets/:id/holds/:hold_id/void", r.walletController.PostVoidHold)

		// PaymentController (PSP connectors: payin via webhook, payout initiation)
		ledger.POST("/payments/webhooks/:psp", r.paymentController.PostWebhook)
		ledger.POST("/payments/payouts", r.paymentController.PostPayout)
		ledger.GET("/payments", r.paymentController.GetPayments)

		// FlowController (Sharia-native orchestration)
		ledger.POST("/flows", r.flowController.PostFlow)
		ledger.GET("/flows", r.flowController.ListFlows)
		ledger.GET("/flows/:id", r.flowController.GetFlow)
		ledger.POST("/flows/:id/trigger", r.flowController.PostTrigger)
		ledger.GET("/flows/:id/instances", r.flowController.GetInstances)
	}

	return engine
}
