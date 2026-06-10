package actions

import (
	"net/http"
	"testing"

	"github.com/amezianechayer/corren/api/middlewares"
	"github.com/amezianechayer/corren/auth"
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// End-to-end: the role matrix enforced against the REAL contract routes,
// not stubs. A readonly key must never be able to execute `sell`.
func TestAuthEnforcedOnContractRoutes(t *testing.T) {
	viper.Set("auth.enabled", true)
	viper.Set("server.http.basic_auth", nil)
	defer viper.Set("auth.enabled", false)

	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("authe2e", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()

			authDir := t.TempDir()
			currentDir := viper.GetString("storage.dir")
			viper.Set("storage.dir", authDir)
			svc, err := auth.NewServiceFromConfig()
			viper.Set("storage.dir", currentDir) // ledger keeps its own dir
			if err != nil {
				t.Fatal(err)
			}

			operatorKey, _, _ := svc.CreateKey("op", auth.RoleOperator, []string{"authe2e"})
			readonlyKey, _, _ := svc.CreateKey("ro", auth.RoleReadonly, nil)

			m := middlewares.NewAuthMiddleware(svc)
			ctl := NewContractController()
			r := gin.New()
			r.Use(m.AuthMiddleware(r))
			grp := r.Group("/:ledger", func(c *gin.Context) { c.Set("ledger", l) })
			grp.POST("/contracts", ctl.PostContract)
			grp.GET("/contracts", ctl.ListContracts)
			grp.POST("/contracts/:id/transitions/:name", ctl.PostTransition)

			// operator creates a contract on its scoped ledger
			code, out := doWithHeaders(t, r, "POST", "/authe2e/contracts", createBody("mur_authe2e"), bearer(operatorKey))
			if code != http.StatusCreated {
				t.Fatalf("operator create: expected 201, got %d: %v", code, out)
			}

			// fund treasury directly (test setup, not via HTTP)
			if _, err := l.Commit([]core.Transaction{{
				Postings: []core.Posting{{Source: core.WORLD, Destination: "@bank:treasury", Asset: "SAR2", Amount: 10000000}},
			}}); err != nil {
				t.Fatal(err)
			}

			// readonly can list…
			code, _ = doWithHeaders(t, r, "GET", "/authe2e/contracts", nil, bearer(readonlyKey))
			if code != http.StatusOK {
				t.Fatalf("readonly list: expected 200, got %d", code)
			}

			// …but can NEVER execute a transition
			code, _ = doWithHeaders(t, r, "POST", "/authe2e/contracts/mur_authe2e/transitions/acquire", nil, bearer(readonlyKey))
			if code != http.StatusForbidden {
				t.Fatalf("readonly acquire: expected 403, got %d", code)
			}
			code, _ = doWithHeaders(t, r, "POST", "/authe2e/contracts/mur_authe2e/transitions/sell", nil, bearer(readonlyKey))
			if code != http.StatusForbidden {
				t.Fatalf("readonly sell: expected 403, got %d", code)
			}

			// operator proceeds normally
			code, out = doWithHeaders(t, r, "POST", "/authe2e/contracts/mur_authe2e/transitions/acquire", nil, bearer(operatorKey))
			if code != http.StatusOK {
				t.Fatalf("operator acquire: expected 200, got %d: %v", code, out)
			}

			// no token at all → 401
			code, _ = doWithHeaders(t, r, "POST", "/authe2e/contracts/mur_authe2e/transitions/sell", nil, nil)
			if code != http.StatusUnauthorized {
				t.Fatalf("anonymous sell: expected 401, got %d", code)
			}
		}),
	)
}
