package actions

import (
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/amezianechayer/corren/ledger"
	"github.com/amezianechayer/corren/wallets"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func withWalletAPI(t *testing.T, f func(r *gin.Engine, l *ledger.Ledger)) {
	t.Helper()
	viper.Set("storage.sqlite.db_name", "wallet_api")
	os.Remove(path.Join(os.TempDir(), "wallet_api_walletapitest.db"))
	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("walletapitest", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()

			wc := NewWalletController()
			r := gin.New()
			grp := r.Group("/:ledger", func(c *gin.Context) { c.Set("ledger", l) })

			grp.POST("/wallets", wc.PostWallet)
			grp.GET("/wallets", wc.ListWallets)
			grp.GET("/wallets/:id", wc.GetWallet)
			grp.GET("/wallets/:id/balances", wc.GetBalances)
			grp.GET("/wallets/:id/transactions", wc.GetTransactions)
			grp.POST("/wallets/:id/credit", wc.PostCredit)
			grp.POST("/wallets/:id/debit", wc.PostDebit)
			grp.POST("/wallets/:id/transfer", wc.PostTransfer)
			grp.POST("/wallets/:id/holds", wc.PostHold)
			grp.GET("/wallets/:id/holds", wc.ListHolds)
			grp.DELETE("/wallets/:id/holds/:hold_id", wc.ReleaseHold)
			grp.POST("/wallets/:id/holds/:hold_id/capture", wc.PostCaptureHold)
			grp.POST("/wallets/:id/holds/:hold_id/void", wc.PostVoidHold)

			f(r, l)
		}),
	)
}

func walletData(out map[string]interface{}) map[string]interface{} {
	return out["data"].(map[string]interface{})
}

func available(t *testing.T, r *gin.Engine, id, asset string) float64 {
	t.Helper()
	code, out := do(t, r, "GET", "/walletapitest/wallets/"+id+"/balances", nil)
	if code != http.StatusOK {
		t.Fatalf("balances: expected 200, got %d: %v", code, out)
	}
	av, _ := walletData(out)["available"].(map[string]interface{})
	if av == nil {
		return 0
	}
	v, _ := av[asset].(float64)
	return v
}

func held(t *testing.T, r *gin.Engine, id, asset string) float64 {
	t.Helper()
	_, out := do(t, r, "GET", "/walletapitest/wallets/"+id+"/balances", nil)
	hl, _ := walletData(out)["held"].(map[string]interface{})
	if hl == nil {
		return 0
	}
	v, _ := hl[asset].(float64)
	return v
}

// Full wallet lifecycle against a real sqlite-backed ledger: this also proves
// the v005 migration runs and balances are read straight from the ledger.
func TestWalletAPIFullFlow(t *testing.T) {
	withWalletAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		const asset = "AED.2"

		// create
		code, out := do(t, r, "POST", "/walletapitest/wallets", map[string]interface{}{
			"owner": "@user:alice", "asset": asset,
		})
		if code != http.StatusCreated {
			t.Fatalf("create: expected 201, got %d: %v", code, out)
		}
		alice := walletData(out)["id"].(string)

		// credit (topup) 1000.00 AED
		if code, out := do(t, r, "POST", "/walletapitest/wallets/"+alice+"/credit", map[string]interface{}{
			"amount": 100000,
		}); code != http.StatusOK {
			t.Fatalf("credit: expected 200, got %d: %v", code, out)
		}
		if got := available(t, r, alice, asset); got != 100000 {
			t.Fatalf("after credit expected available 100000, got %v", got)
		}

		// debit 200.00 out (cash-out to @world)
		if code, out := do(t, r, "POST", "/walletapitest/wallets/"+alice+"/debit", map[string]interface{}{
			"amount": 20000,
		}); code != http.StatusOK {
			t.Fatalf("debit: expected 200, got %d: %v", code, out)
		}
		if got := available(t, r, alice, asset); got != 80000 {
			t.Fatalf("after debit expected available 80000, got %v", got)
		}

		// overdraw → 422 ERR_INSUFFICIENT_FUNDS
		code, out = do(t, r, "POST", "/walletapitest/wallets/"+alice+"/debit", map[string]interface{}{
			"amount": 999999,
		})
		if code != http.StatusUnprocessableEntity || out["error"] != wallets.ErrInsufficientFunds {
			t.Fatalf("overdraw: expected 422 ERR_INSUFFICIENT_FUNDS, got %d: %v", code, out)
		}

		// hold 300.00 → available 500.00, held 300.00
		code, out = do(t, r, "POST", "/walletapitest/wallets/"+alice+"/holds", map[string]interface{}{
			"amount": 30000, "reason": "pending order", "expires_at": "2026-12-31T00:00:00Z",
		})
		if code != http.StatusCreated {
			t.Fatalf("hold: expected 201, got %d: %v", code, out)
		}
		holdID := walletData(out)["id"].(string)
		if av, hl := available(t, r, alice, asset), held(t, r, alice, asset); av != 50000 || hl != 30000 {
			t.Fatalf("after hold expected available 50000 / held 30000, got %v / %v", av, hl)
		}

		// capture hold → held back to 0
		if code, out := do(t, r, "POST", "/walletapitest/wallets/"+alice+"/holds/"+holdID+"/capture", map[string]interface{}{}); code != http.StatusOK {
			t.Fatalf("capture: expected 200, got %d: %v", code, out)
		}
		if hl := held(t, r, alice, asset); hl != 0 {
			t.Fatalf("after capture expected held 0, got %v", hl)
		}

		// P2P: create Bob, transfer 100.00 from Alice to Bob's main account
		_, out = do(t, r, "POST", "/walletapitest/wallets", map[string]interface{}{"owner": "@user:bob", "asset": asset})
		bob := walletData(out)["id"].(string)
		if code, out := do(t, r, "POST", "/walletapitest/wallets/"+alice+"/debit", map[string]interface{}{
			"amount": 10000, "destination": wallets.MainAccount(bob),
		}); code != http.StatusOK {
			t.Fatalf("p2p debit: expected 200, got %d: %v", code, out)
		}
		if got := available(t, r, bob, asset); got != 10000 {
			t.Fatalf("after P2P expected Bob available 10000, got %v", got)
		}
	})
}

// Exercises the spec's exact surface: GET /:id returns wallet+balance, the
// /transfer endpoint, DELETE to release a hold, and /transactions history.
func TestWalletAPISpecSurface(t *testing.T) {
	withWalletAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		const asset = "AED.2"

		_, out := do(t, r, "POST", "/walletapitest/wallets", map[string]interface{}{"owner": "@user:alice", "asset": asset})
		alice := walletData(out)["id"].(string)
		_, out = do(t, r, "POST", "/walletapitest/wallets", map[string]interface{}{"owner": "@user:bob", "asset": asset})
		bob := walletData(out)["id"].(string)
		do(t, r, "POST", "/walletapitest/wallets/"+alice+"/credit", map[string]interface{}{"amount": 100000})

		// GET /wallets/:id returns wallet + balance
		code, out := do(t, r, "GET", "/walletapitest/wallets/"+alice, nil)
		if code != http.StatusOK {
			t.Fatalf("get wallet: %d: %v", code, out)
		}
		d := walletData(out)
		if d["owner"] != "@user:alice" {
			t.Fatalf("expected owner in wallet response, got %v", d)
		}
		bals, ok := d["balances"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected balances in GET /:id, got %v", d)
		}
		av, _ := bals["available"].(map[string]interface{})
		if av[asset].(float64) != 100000 {
			t.Fatalf("expected available 100000 in GET /:id, got %v", bals)
		}

		// transfer 300.00 Alice -> Bob via the /transfer endpoint
		if code, out := do(t, r, "POST", "/walletapitest/wallets/"+alice+"/transfer", map[string]interface{}{
			"to": bob, "amount": 30000,
		}); code != http.StatusOK {
			t.Fatalf("transfer: %d: %v", code, out)
		}
		if got := available(t, r, bob, asset); got != 30000 {
			t.Fatalf("after transfer expected Bob available 30000, got %v", got)
		}

		// transfer to a non-existent wallet → 404
		if code, _ := do(t, r, "POST", "/walletapitest/wallets/"+alice+"/transfer", map[string]interface{}{
			"to": "wlt_ghost", "amount": 1000,
		}); code != http.StatusNotFound {
			t.Fatalf("transfer to ghost: expected 404, got %d", code)
		}

		// hold then DELETE to release
		code, out = do(t, r, "POST", "/walletapitest/wallets/"+alice+"/holds", map[string]interface{}{
			"amount": 20000, "reason": "escrow", "expires_at": "2026-12-31T00:00:00Z",
		})
		if code != http.StatusCreated {
			t.Fatalf("hold: %d: %v", code, out)
		}
		hold := walletData(out)
		if hold["reason"] != "escrow" || hold["expires_at"] != "2026-12-31T00:00:00Z" {
			t.Fatalf("hold should carry reason+expires_at, got %v", hold)
		}
		holdID := hold["id"].(string)
		if held(t, r, alice, asset) != 20000 {
			t.Fatalf("expected held 20000 after hold")
		}
		if code, out := do(t, r, "DELETE", "/walletapitest/wallets/"+alice+"/holds/"+holdID, nil); code != http.StatusOK {
			t.Fatalf("release hold: %d: %v", code, out)
		}
		if held(t, r, alice, asset) != 0 {
			t.Fatalf("expected held 0 after release")
		}

		// transaction history is non-empty
		code, out = do(t, r, "GET", "/walletapitest/wallets/"+alice+"/transactions", nil)
		if code != http.StatusOK {
			t.Fatalf("transactions: %d: %v", code, out)
		}
		cur, _ := out["cursor"].(map[string]interface{})
		data, _ := cur["data"].([]interface{})
		if len(data) == 0 {
			t.Fatalf("expected non-empty transaction history, got %v", out)
		}
	})
}
