package api

import (
	"context"
	_ "embed"
	"strings"
	"time"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/ledger/query"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type HttpAPI struct {
	addr   string
	engine *gin.Engine
}

func NewHttpAPI(lc fx.Lifecycle, resolver *ledger.Resolver) *HttpAPI {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(gin.Recovery())

	if auth := viper.Get("server.http.basic_auth"); auth != nil {
		segment := strings.Split(auth.(string), ":")
		r.Use(gin.BasicAuth(gin.Accounts{
			segment[0]: segment[1],
		}))
	}

	if keys := viper.GetStringMapString("server.http.api_keys"); len(keys) > 0 {
		r.Use(func(c *gin.Context) {
			key := c.GetHeader("X-API-Key")
			if key == "" {
				c.AbortWithStatusJSON(401, gin.H{"ok": false, "err": "missing X-API-Key header"})
				return
			}
			if role, ok := keys[key]; ok {
				c.Set("api_key_role", role)
				c.Next()
				return
			}
			c.AbortWithStatusJSON(401, gin.H{"ok": false, "err": "invalid API key"})
		})
	}

	r.Use(func(c *gin.Context) {
		name := c.Param("ledger")

		if name == "" {
			return
		}

		l, err := resolver.GetLedger(name)

		if err != nil {
			c.JSON(400, gin.H{
				"ok":  false,
				"err": err.Error(),
			})
		}

		c.Set("ledger", l)
	})

	r.GET("/_healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true, "status": "healthy"})
	})

	r.GET("/_info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"server":  "corren-ledger",
			"version": "1.0.0-alpha.1",
			"config": gin.H{
				"storage": gin.H{
					"driver": viper.Get("storage.driver"),
				},
				"ledgers": viper.Get("ledgers"),
			},
		})
	})

	r.GET("/:ledger/stats", func(c *gin.Context) {
		l, _ := c.Get("ledger")

		stats, err := l.(*ledger.Ledger).Stats()

		c.JSON(200, gin.H{
			"ok":    err == nil,
			"stats": stats,
		})
	})

	r.GET("/:ledger/transactions", func(c *gin.Context) {
		l, _ := c.Get("ledger")

		cursor, err := l.(*ledger.Ledger).FindTransactions(
			query.After(c.Query("after")),
			query.Account(c.Query("account")),
		)

		c.JSON(200, gin.H{
			"ok":     err == nil,
			"cursor": cursor,
			"err":    err,
		})
	})

	r.POST("/:ledger/transactions", func(c *gin.Context) {
		l, _ := c.Get("ledger")

		var t core.Transaction
		c.ShouldBind(&t)

		err := l.(*ledger.Ledger).Commit([]core.Transaction{t})

		c.JSON(200, gin.H{
			"ok": err == nil,
		})
	})

	r.POST("/:ledger/script", func(c *gin.Context) {
		l, _ := c.Get("ledger")

		var script core.Script
		c.ShouldBind(&script)

		err := l.(*ledger.Ledger).Execute(script)

		res := gin.H{
			"ok": err == nil,
		}

		if err != nil {
			res["err"] = err.Error()
		}

		c.JSON(200, res)
	})

	r.GET("/:ledger/accounts", func(c *gin.Context) {
		l, _ := c.Get("ledger")

		cursor, err := l.(*ledger.Ledger).FindAccounts(
			query.After(c.Query("after")),
		)

		res := gin.H{
			"ok":     err == nil,
			"cursor": cursor,
		}

		if err != nil {
			res["err"] = err.Error()
		}

		c.JSON(200, res)
	})

	r.GET("/:ledger/accounts/:address", func(c *gin.Context) {
		l, _ := c.Get("ledger")

		acc, err := l.(*ledger.Ledger).GetAccount(c.Param("address"))

		res := gin.H{
			"ok":      err == nil,
			"account": acc,
		}

		if err != nil {
			res["err"] = err.Error()
		}

		c.JSON(200, res)
	})

	r.GET("/:ledger/assets", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		assets, err := l.(*ledger.Ledger).FindAssets()
		res := gin.H{"ok": err == nil, "assets": assets}
		if err != nil {
			res["err"] = err.Error()
		}
		c.JSON(200, res)
	})

	r.POST("/:ledger/assets", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		var a core.AssetEntry
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(400, gin.H{"ok": false, "err": err.Error()})
			return
		}
		a.CreatedAt = time.Now().Format(time.RFC3339)
		err := l.(*ledger.Ledger).SaveAsset(a)
		res := gin.H{"ok": err == nil}
		if err != nil {
			res["err"] = err.Error()
		}
		c.JSON(200, res)
	})

	r.GET("/:ledger/assets/:id", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		a, err := l.(*ledger.Ledger).GetAsset(c.Param("id"))
		if err != nil {
			c.JSON(500, gin.H{"ok": false, "err": err.Error()})
			return
		}
		if a == nil {
			c.JSON(404, gin.H{"ok": false, "err": "asset not found"})
			return
		}
		c.JSON(200, gin.H{"ok": true, "asset": a})
	})

	r.GET("/:ledger/contracts", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		cursor, err := l.(*ledger.Ledger).FindContracts(
			query.After(c.Query("after")),
		)
		res := gin.H{"ok": err == nil, "cursor": cursor}
		if err != nil {
			res["err"] = err.Error()
		}
		c.JSON(200, res)
	})

	r.POST("/:ledger/contracts", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		var contract core.ShariaContract
		if err := c.ShouldBindJSON(&contract); err != nil {
			c.JSON(400, gin.H{"ok": false, "err": err.Error()})
			return
		}
		now := time.Now().Format(time.RFC3339)
		contract.CreatedAt = now
		contract.UpdatedAt = now
		if contract.Status == "" {
			contract.Status = core.ContractPending
		}
		if contract.AaoifiFAS == "" {
			contract.AaoifiFAS = core.AaoifiFAS[contract.Type]
		}
		err := l.(*ledger.Ledger).SaveContract(contract)
		if err != nil {
			c.JSON(200, gin.H{"ok": false, "err": err.Error()})
			return
		}
		c.JSON(200, gin.H{"ok": true, "contract": contract})
	})

	r.GET("/:ledger/contracts/:id", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		contract, err := l.(*ledger.Ledger).GetContract(c.Param("id"))
		if err != nil {
			c.JSON(500, gin.H{"ok": false, "err": err.Error()})
			return
		}
		if contract == nil {
			c.JSON(404, gin.H{"ok": false, "err": "contract not found"})
			return
		}
		c.JSON(200, gin.H{"ok": true, "contract": contract})
	})

	r.PATCH("/:ledger/contracts/:id/status", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		var body struct {
			Status core.ContractStatus `json:"status"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(400, gin.H{"ok": false, "err": err.Error()})
			return
		}
		err := l.(*ledger.Ledger).UpdateContractStatus(c.Param("id"), body.Status)
		res := gin.H{"ok": err == nil}
		if err != nil {
			res["err"] = err.Error()
		}
		c.JSON(200, res)
	})

	r.GET("/:ledger/certificates", func(c *gin.Context) {
		l, _ := c.Get("ledger")
		contractID := c.Query("contract_id")
		certs, err := l.(*ledger.Ledger).FindCertificates(contractID)
		res := gin.H{"ok": err == nil, "certificates": certs}
		if err != nil {
			res["err"] = err.Error()
		}
		c.JSON(200, res)
	})

	h := &HttpAPI{
		engine: r,
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

func (h *HttpAPI) Start() {
	h.engine.Run(h.addr)
}
