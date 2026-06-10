package actions

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/amezianechayer/corren/config"
	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	gin.SetMode(gin.TestMode)
	config.Init()
	viper.Set("storage.dir", os.TempDir())
	viper.Set("storage.sqlite.db_name", "sharia_api")
	os.Remove(path.Join(os.TempDir(), "sharia_api_apitest.db"))
	os.Exit(m.Run())
}

func withAPI(t *testing.T, f func(r *gin.Engine, l *ledger.Ledger)) {
	t.Helper()
	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("apitest", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()

			ctl := NewContractController()
			r := gin.New()
			grp := r.Group("/:ledger", func(c *gin.Context) { c.Set("ledger", l) })
			grp.POST("/contracts", ctl.PostContract)
			grp.GET("/contracts", ctl.ListContracts)
			grp.GET("/contracts/:id", ctl.GetContract)
			grp.POST("/contracts/:id/transitions/:name", ctl.PostTransition)
			grp.GET("/contracts/:id/schedule", ctl.GetSchedule)
			grp.GET("/contracts/:id/audit", ctl.GetAudit)

			f(r, l)
		}),
	)
}

func do(t *testing.T, r *gin.Engine, method, url string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	return doWithHeaders(t, r, method, url, body, nil)
}

func doWithHeaders(t *testing.T, r *gin.Engine, method, url string, body interface{}, headers map[string]string) (int, map[string]interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	out := map[string]interface{}{}
	if len(w.Body.Bytes()) > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("invalid JSON response (%d): %s", w.Code, w.Body.String())
		}
	}
	return w.Code, out
}

func createBody(id string) map[string]interface{} {
	return map[string]interface{}{
		"type": "murabaha",
		"id":   id,
		"params": map[string]interface{}{
			"asset_code":   "VHCL42A",
			"cost":         map[string]interface{}{"asset": "SAR2", "amount": 10000000},
			"markup":       map[string]interface{}{"asset": "SAR2", "amount": 1000000},
			"client":       "@client:api",
			"supplier":     "@supplier:api",
			"installments": 24,
			"first_due":    "2026-07-01T00:00:00Z",
			"period_days":  30,
		},
	}
}

func TestContractAPIFullFlow(t *testing.T) {
	withAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		const id = "mur_api_flow"

		// create → 201
		code, out := do(t, r, "POST", "/apitest/contracts", createBody(id))
		if code != http.StatusCreated {
			t.Fatalf("create: expected 201, got %d: %v", code, out)
		}
		data := out["data"].(map[string]interface{})
		contract := data["contract"].(map[string]interface{})
		if contract["state"] != "PROMISE" {
			t.Fatalf("expected PROMISE, got %v", contract["state"])
		}
		if schedule := data["schedule"].([]interface{}); len(schedule) != 24 {
			t.Fatalf("expected 24 installments, got %d", len(schedule))
		}

		// sell before acquire → 422 referenced rejection (I-1)
		code, out = do(t, r, "POST", "/apitest/contracts/"+id+"/transitions/sell", nil)
		if code != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d: %v", code, out)
		}
		if out["error"] != "ERR_SHARIA_VIOLATION" || out["standard_ref"] != "AAOIFI-SS-8" {
			t.Fatalf("expected referenced sharia violation, got %v", out)
		}

		// fund treasury, acquire → 200
		if _, err := l.Commit([]core.Transaction{{
			Postings: []core.Posting{{Source: core.WORLD, Destination: "@bank:treasury", Asset: "SAR2", Amount: 10000000}},
		}}); err != nil {
			t.Fatal(err)
		}
		code, out = do(t, r, "POST", "/apitest/contracts/"+id+"/transitions/acquire", nil)
		if code != http.StatusOK {
			t.Fatalf("acquire: expected 200, got %d: %v", code, out)
		}
		data = out["data"].(map[string]interface{})
		if data["new_state"] != "ACQUIRED" {
			t.Fatalf("expected ACQUIRED, got %v", data["new_state"])
		}

		// sell → 200 SOLD
		code, out = do(t, r, "POST", "/apitest/contracts/"+id+"/transitions/sell", nil)
		if code != http.StatusOK {
			t.Fatalf("sell: expected 200, got %d: %v", code, out)
		}

		// pay installment 1 → 200
		if _, err := l.Commit([]core.Transaction{{
			Postings: []core.Posting{{Source: core.WORLD, Destination: "@client:api", Asset: "SAR2", Amount: 500000}},
		}}); err != nil {
			t.Fatal(err)
		}
		code, out = do(t, r, "POST", "/apitest/contracts/"+id+"/transitions/pay_installment", map[string]interface{}{"seq": 1})
		if code != http.StatusOK {
			t.Fatalf("pay: expected 200, got %d: %v", code, out)
		}

		// replay → 409 ERR_DUPLICATE
		code, out = do(t, r, "POST", "/apitest/contracts/"+id+"/transitions/pay_installment", map[string]interface{}{"seq": 1})
		if code != http.StatusConflict || out["error"] != "ERR_DUPLICATE" {
			t.Fatalf("expected 409 ERR_DUPLICATE, got %d: %v", code, out)
		}

		// penalty to non-charity → 422 SS3
		code, out = do(t, r, "POST", "/apitest/contracts/"+id+"/transitions/late_penalty",
			map[string]interface{}{"seq": 2, "amount": 20000, "destination": "@bank:income:murabaha"})
		if code != http.StatusUnprocessableEntity || out["standard_ref"] != "AAOIFI-SS-3" {
			t.Fatalf("expected 422 SS-3, got %d: %v", code, out)
		}

		// schedule → 200, seq 1 paid
		code, out = do(t, r, "GET", "/apitest/contracts/"+id+"/schedule", nil)
		if code != http.StatusOK {
			t.Fatalf("schedule: expected 200, got %d", code)
		}
		items := out["data"].([]interface{})
		if items[0].(map[string]interface{})["status"] != "paid" {
			t.Fatalf("expected seq 1 paid, got %v", items[0])
		}

		// audit with verify → chain_valid true and denied events present
		code, out = do(t, r, "GET", "/apitest/contracts/"+id+"/audit?verify=true", nil)
		if code != http.StatusOK {
			t.Fatalf("audit: expected 200, got %d", code)
		}
		data = out["data"].(map[string]interface{})
		if data["chain_valid"] != true {
			t.Fatalf("expected chain_valid true, got %v", data["chain_valid"])
		}
		denied := false
		for _, ev := range data["events"].([]interface{}) {
			if ev.(map[string]interface{})["decision"] == "denied" {
				denied = true
			}
		}
		if !denied {
			t.Fatal("expected at least one denied event in audit")
		}

		// get contract → 200
		code, out = do(t, r, "GET", "/apitest/contracts/"+id, nil)
		if code != http.StatusOK {
			t.Fatalf("get: expected 200, got %d", code)
		}

		// list → 200
		code, out = do(t, r, "GET", "/apitest/contracts", nil)
		if code != http.StatusOK {
			t.Fatalf("list: expected 200, got %d", code)
		}
	})
}

func TestContractAPIErrors(t *testing.T) {
	withAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		// invalid params → 400
		body := createBody("mur_api_bad")
		body["params"].(map[string]interface{})["installments"] = 0
		code, out := do(t, r, "POST", "/apitest/contracts", body)
		if code != http.StatusBadRequest || out["error"] != "ERR_INVALID_PARAMS" {
			t.Fatalf("expected 400 ERR_INVALID_PARAMS, got %d: %v", code, out)
		}

		// unknown contract → 404
		code, out = do(t, r, "GET", "/apitest/contracts/mur_api_unknown", nil)
		if code != http.StatusNotFound || out["error"] != "ERR_NOT_FOUND" {
			t.Fatalf("expected 404 ERR_NOT_FOUND, got %d: %v", code, out)
		}

		// unknown transition on existing contract → 409
		if code, _ := do(t, r, "POST", "/apitest/contracts", createBody("mur_api_err")); code != http.StatusCreated {
			t.Fatal("setup create failed")
		}
		code, out = do(t, r, "POST", "/apitest/contracts/mur_api_err/transitions/liquidate", nil)
		if code != http.StatusConflict || out["error"] != "ERR_INVALID_TRANSITION" {
			t.Fatalf("expected 409 ERR_INVALID_TRANSITION, got %d: %v", code, out)
		}
	})
}
