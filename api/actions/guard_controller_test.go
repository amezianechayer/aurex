package actions

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/guard"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func withGuardAPI(t *testing.T, f func(r *gin.Engine, l *ledger.Ledger)) {
	t.Helper()
	viper.Set("storage.sqlite.db_name", "guard_api")
	os.Remove(path.Join(os.TempDir(), "guard_api_guardapitest.db"))
	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("guardapitest", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()

			gctl := NewGuardController()
			tctl := NewTransactionController()

			r := gin.New()
			grp := r.Group("/:ledger", func(c *gin.Context) { c.Set("ledger", l) })

			// transaction routes (for the deny-via-/transactions test)
			grp.POST("/transactions", tctl.PostTransaction)
			grp.GET("/accounts/:address", func(c *gin.Context) {
				acct, err := l.GetAccount(c.Param("address"))
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"ok": true, "data": acct})
			})

			// guard routes
			grp.POST("/guard/rules", gctl.PostRule)
			grp.GET("/guard/rules", gctl.ListRules)
			grp.GET("/guard/rules/:id", gctl.GetRule)
			grp.PATCH("/guard/rules/:id", gctl.PatchRule)
			grp.DELETE("/guard/rules/:id", gctl.DeleteRule)
			grp.GET("/guard/events", gctl.ListEvents)

			f(r, l)
		}),
	)
}

func amountCapRuleBody(withStandardRef bool) map[string]interface{} {
	params, _ := json.Marshal(map[string]interface{}{
		"scope": "@client:*",
		"asset": "SAR2",
		"max":   500000,
		"basis": "posting",
	})
	body := map[string]interface{}{
		"kind":    "amount_cap",
		"action":  "deny",
		"reason":  "cap single posting to 5000 SAR",
		"enabled": true,
		"params":  json.RawMessage(params),
	}
	if withStandardRef {
		body["standard_ref"] = "AAOIFI-SS-17"
	}
	return body
}

func TestGuardRuleCRUD(t *testing.T) {
	withGuardAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		// POST a valid deny rule (with standard_ref) → 201
		code, out := do(t, r, "POST", "/guardapitest/guard/rules", amountCapRuleBody(true))
		if code != http.StatusCreated {
			t.Fatalf("create rule: expected 201, got %d: %v", code, out)
		}
		data := out["data"].(map[string]interface{})
		ruleID, ok := data["id"].(string)
		if !ok || ruleID == "" {
			t.Fatalf("expected a rule id in response, got %v", data)
		}
		if data["kind"] != "amount_cap" {
			t.Fatalf("expected kind amount_cap, got %v", data["kind"])
		}

		// POST a deny rule WITHOUT standard_ref → 400 ERR_INVALID_RULE
		code, out = do(t, r, "POST", "/guardapitest/guard/rules", amountCapRuleBody(false))
		if code != http.StatusBadRequest {
			t.Fatalf("missing std_ref: expected 400, got %d: %v", code, out)
		}
		if out["error"] != "ERR_INVALID_RULE" {
			t.Fatalf("expected ERR_INVALID_RULE, got %v", out["error"])
		}

		// GET /guard/rules → 200, list non-empty
		code, out = do(t, r, "GET", "/guardapitest/guard/rules", nil)
		if code != http.StatusOK {
			t.Fatalf("list rules: expected 200, got %d: %v", code, out)
		}
		list, ok := out["data"].([]interface{})
		if !ok || len(list) == 0 {
			t.Fatalf("expected non-empty rule list, got %v", out)
		}

		// GET /guard/rules/:id → 200
		code, out = do(t, r, "GET", "/guardapitest/guard/rules/"+ruleID, nil)
		if code != http.StatusOK {
			t.Fatalf("get rule: expected 200, got %d: %v", code, out)
		}
		if out["data"].(map[string]interface{})["id"] != ruleID {
			t.Fatalf("get rule: expected id %s, got %v", ruleID, out)
		}

		// PATCH /guard/rules/:id → 200 (update reason)
		code, out = do(t, r, "PATCH", "/guardapitest/guard/rules/"+ruleID, map[string]interface{}{
			"reason": "updated cap reason",
		})
		if code != http.StatusOK {
			t.Fatalf("patch rule: expected 200, got %d: %v", code, out)
		}
		if out["data"].(map[string]interface{})["reason"] != "updated cap reason" {
			t.Fatalf("patch rule: reason not updated, got %v", out["data"])
		}

		// DELETE /guard/rules/:id → 200
		code, out = do(t, r, "DELETE", "/guardapitest/guard/rules/"+ruleID, nil)
		if code != http.StatusOK {
			t.Fatalf("delete rule: expected 200, got %d: %v", code, out)
		}
		if out["data"].(map[string]interface{})["deleted"] != true {
			t.Fatalf("expected deleted true, got %v", out)
		}

		// Confirm rule is gone → 404
		code, out = do(t, r, "GET", "/guardapitest/guard/rules/"+ruleID, nil)
		if code != http.StatusNotFound {
			t.Fatalf("expected 404 after delete, got %d: %v", code, out)
		}
	})
}

func TestGuardDenyViaTransactions(t *testing.T) {
	withGuardAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		// Fund a client account
		if _, err := l.Commit([]core.Transaction{{
			Postings: []core.Posting{{
				Source:      core.WORLD,
				Destination: "@client:guard-test",
				Asset:       "SAR2",
				Amount:      2000000,
			}},
		}}); err != nil {
			t.Fatalf("fund client: %v", err)
		}

		// Create a deny rule: cap single posting from @client:* to 500000 SAR2
		code, out := do(t, r, "POST", "/guardapitest/guard/rules", amountCapRuleBody(true))
		if code != http.StatusCreated {
			t.Fatalf("create deny rule: expected 201, got %d: %v", code, out)
		}
		ruleID := out["data"].(map[string]interface{})["id"].(string)

		// Attempt a transfer exceeding the cap (amount 600000 > max 500000)
		// This should be blocked. The transaction controller maps Commit errors to 500.
		code, out = do(t, r, "POST", "/guardapitest/transactions", map[string]interface{}{
			"postings": []map[string]interface{}{
				{
					"source":      "@client:guard-test",
					"destination": "@bank:treasury",
					"asset":       "SAR2",
					"amount":      600000,
				},
			},
		})
		// A guard deny is an expected policy rejection: the transaction controller
		// surfaces it as 422 (not 500) carrying the rule's code and standard_ref.
		if code != http.StatusUnprocessableEntity {
			t.Fatalf("blocked transfer: expected 422, got %d: %v", code, out)
		}
		if out["error"] != guard.ErrGuardDenied {
			t.Fatalf("expected error %s, got %v", guard.ErrGuardDenied, out["error"])
		}
		if out["standard_ref"] != "AAOIFI-SS-17" {
			t.Fatalf("expected standard_ref echoed, got %v", out)
		}

		// Verify balance unchanged: client should still have 2000000
		code, out = do(t, r, "GET", "/guardapitest/accounts/@client:guard-test", nil)
		if code != http.StatusOK {
			t.Fatalf("get account: expected 200, got %d: %v", code, out)
		}
		acct := out["data"].(map[string]interface{})
		balances, ok := acct["balances"].(map[string]interface{})
		if !ok {
			// try volumes
			if volumes, ok2 := acct["volumes"].(map[string]interface{}); ok2 {
				sarVol, ok3 := volumes["SAR2"].(map[string]interface{})
				if ok3 {
					input := int64(sarVol["input"].(float64))
					output := int64(sarVol["output"].(float64))
					balance := input - output
					if balance != 2000000 {
						t.Fatalf("expected client balance 2000000 after blocked transfer, got %d", balance)
					}
				}
			}
		} else {
			sarBalance, ok := balances["SAR2"]
			if ok {
				if sarBalance.(float64) != 2000000 {
					t.Fatalf("expected client balance 2000000 after blocked transfer, got %v", sarBalance)
				}
			}
		}

		// GET /guard/events → 200, the deny event is present
		code, out = do(t, r, "GET", "/guardapitest/guard/events", nil)
		if code != http.StatusOK {
			t.Fatalf("list events: expected 200, got %d: %v", code, out)
		}
		events, ok := out["data"].([]interface{})
		if !ok {
			t.Fatalf("expected events list, got %v", out)
		}
		foundDeny := false
		for _, ev := range events {
			evMap := ev.(map[string]interface{})
			if evMap["action"] == "deny" && evMap["rule_id"] == ruleID {
				foundDeny = true
			}
		}
		if !foundDeny {
			t.Fatalf("expected a deny guard event for rule %s, events: %v", ruleID, events)
		}

		// Clean up: delete the rule
		code, out = do(t, r, "DELETE", "/guardapitest/guard/rules/"+ruleID, nil)
		if code != http.StatusOK {
			t.Fatalf("delete rule: expected 200, got %d: %v", code, out)
		}
	})
}
