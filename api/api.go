package api

import (
	"context"
	_ "embed"

	"github.com/amezianechayer/corren/api/actions"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Module exported for initializing application
var Module = fx.Options(
	actions.Module,
)

type API struct {
	addr   string
	engine *gin.Engine
}

func NewAPI(
	lc fx.Lifecycle,
	resolver *ledger.Resolver,
	configController *actions.ConfigController, //todo: use fx
	ledgerController *actions.LedgerController, //todo: use fx
	scriptController *actions.ScriptController, //todo: use fx
	accountController *actions.AccountController, //todo: use fx
	transactionController *actions.TransactionController, //todo: use fx
) *API {
	gin.SetMode(gin.ReleaseMode)

	cc := cors.DefaultConfig()
	cc.AllowAllOrigins = true
	cc.AllowCredentials = true
	cc.AddAllowHeaders("authorization")

	//todo: use fx
	router := NewRoutes(
		cc,
		resolver,
		configController,
		ledgerController,
		scriptController,
		accountController,
		transactionController,
	)

	h := &API{
		engine: router, //todo: use fx
		addr:   viper.GetString("server.http.bind_address"),
	}

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			go h.Start()
			return nil
		},
	})

	return h
}

func (h *API) Start() {
	h.engine.Run(h.addr)
}
