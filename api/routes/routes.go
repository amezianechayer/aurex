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
	}

	return engine
}
