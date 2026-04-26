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
	configController      *actions.ConfigController
	ledgerController      *actions.LedgerController
	scriptController      *actions.ScriptController
	accountController     *actions.AccountController
	transactionController *actions.TransactionController
}

// NewRoutes -
func NewRoutes(
	resolver *ledger.Resolver,
	configController *actions.ConfigController,
	ledgerController *actions.LedgerController,
	scriptController *actions.ScriptController,
	accountController *actions.AccountController,
	transactionController *actions.TransactionController,
) *Routes {
	return &Routes{
		resolver:              resolver,
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
		middlewares.AuthMiddleware(engine),
	)

	// API Routes
	engine.GET("/_info", r.configController.GetInfo)

	ledgerGroup := engine.Group("/:ledger", middlewares.LedgerMiddleware(r.resolver))
	{
		// LedgerController
		ledgerGroup.GET("/stats", r.ledgerController.GetStats)

		// TransactionController
		ledgerGroup.GET("/transactions", r.transactionController.GetTransactions)
		ledgerGroup.POST("/transactions", r.transactionController.PostTransaction)
		ledgerGroup.POST("/transactions/:transactionId/revert", r.transactionController.RevertTransaction)
		ledgerGroup.GET("/transactions/:transactionId/metadata", r.transactionController.GetTransactionMetadata)

		// AccountController
		ledgerGroup.GET("/accounts", r.accountController.GetAccounts)
		ledgerGroup.GET("/accounts/:accountId", r.accountController.GetAddress)
		ledgerGroup.GET("/accounts/:accountId/metadata", r.accountController.GetAccountMetadata)

		// ScriptController
		ledgerGroup.POST("/script", r.scriptController.PostScript)
	}

	return engine
}
