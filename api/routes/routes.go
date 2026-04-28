package routes

import (
	"github.com/amezianechayer/corren/api/actions"
	"github.com/amezianechayer/corren/api/middlewares"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-contrib/cors"
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
	}
}

// Engine -
func (r *Routes) Engine(cc cors.Config) *gin.Engine {
	engine := gin.Default()

	// Default Middlewares
	engine.Use(
		cors.New(cc),
		gin.Recovery(),
		r.authMiddleware.AuthMiddleware(engine),
	)

	// API Routes
	engine.GET("/_info", r.configController.GetInfo)

	ledger := engine.Group("/:ledger", r.ledgerMiddleware.LedgerMiddleware())
	{
		// LedgerController
		ledger.GET("/stats", r.ledgerController.GetStats)

		// TransactionController
		ledger.GET("/transactions", r.transactionController.GetTransactions)
		ledger.POST("/transactions", r.transactionController.PostTransaction)
		ledger.POST("/transactions/:transactionId/revert", r.transactionController.RevertTransaction)
		ledger.GET("/transactions/:transactionId/metadata", r.transactionController.GetTransactionMetadata)

		// AccountController
		ledger.GET("/accounts", r.accountController.GetAccounts)
		ledger.GET("/accounts/:accountId", r.accountController.GetAddress)
		ledger.GET("/accounts/:accountId/metadata", r.accountController.GetAccountMetadata)

		// ScriptController
		ledger.POST("/script", r.scriptController.PostScript)
	}

	return engine
}
