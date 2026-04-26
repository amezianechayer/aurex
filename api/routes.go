package api

import (
	"github.com/amezianechayer/corren/api/actions"
	"github.com/amezianechayer/corren/api/middlewares"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewRoutes -
func NewRoutes(
	cc cors.Config,
	resolver *ledger.Resolver,
	configController *actions.ConfigController,
	ledgerController *actions.LedgerController,
	scriptController *actions.ScriptController,
	accountController *actions.AccountController,
	transactionController *actions.TransactionController,
) *gin.Engine {
	engine := gin.Default()

	// Default Middlewares
	engine.Use(
		cors.New(cc),
		gin.Recovery(),
		middlewares.AuthMiddleware(engine),
	)

	// API Routes
	engine.GET("/_info", configController.GetInfo)

	ledgerGroup := engine.Group("/:ledger", middlewares.LedgerMiddleware(resolver))
	{
		// LedgerController
		ledgerGroup.GET("/stats", ledgerController.GetStats)

		// TransactionController
		ledgerGroup.GET("/transactions", transactionController.GetTransactions)
		ledgerGroup.POST("/transactions", transactionController.PostTransaction)
		ledgerGroup.POST("/transactions/:transactionId/revert", transactionController.RevertTransaction)
		ledgerGroup.GET("/transactions/:transactionId/metadata", transactionController.GetTransactionMetadata)

		// AccountController
		ledgerGroup.GET("/accounts", accountController.GetAccounts)
		ledgerGroup.GET("/accounts/:accountId", accountController.GetAddress)
		ledgerGroup.GET("/accounts/:accountId/metadata", accountController.GetAccountMetadata)

		// ScriptController
		ledgerGroup.POST("/script", scriptController.PostScript)
	}

	return engine
}
