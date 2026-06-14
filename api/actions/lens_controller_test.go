package actions

import (
	"net/http"
	"testing"

	"github.com/amezianechayer/corren/core"
	"github.com/amezianechayer/corren/ledger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// withLensAPI is a minimal test harness for the LensController.
// It mirrors withAPI but registers only lens routes and uses a separate
// ledger ("lenstest") to avoid interfering with contract test data.
func withLensAPI(t *testing.T, f func(r *gin.Engine, l *ledger.Ledger)) {
	t.Helper()
	fx.New(
		fx.NopLogger,
		fx.Provide(func(lc fx.Lifecycle) (*ledger.Ledger, error) {
			return ledger.NewLedger("lenstest", lc)
		}),
		fx.Invoke(func(l *ledger.Ledger) {
			defer l.Close()

			ctl := NewLensController()
			r := gin.New()
			grp := r.Group("/:ledger", func(c *gin.Context) { c.Set("ledger", l) })
			grp.GET("/lens/overview", ctl.Overview)
			grp.GET("/lens/flows", ctl.Flows)
			grp.GET("/lens/rollup", ctl.Rollup)
			grp.GET("/lens/timeseries", ctl.TimeSeries)

			f(r, l)
		}),
	)
}

// seedLensData commits two transactions:
//
//	WORLD → @client:a  1000 DZD.2
//	@client:a → @bank:treasury  400 DZD.2
func seedLensData(t *testing.T, l *ledger.Ledger) {
	t.Helper()
	if _, err := l.Commit([]core.Transaction{{
		Postings: []core.Posting{{
			Source:      core.WORLD,
			Destination: "@client:a",
			Asset:       "DZD.2",
			Amount:      1000,
		}},
	}}); err != nil {
		t.Fatal("seed tx1:", err)
	}
	if _, err := l.Commit([]core.Transaction{{
		Postings: []core.Posting{{
			Source:      "@client:a",
			Destination: "@bank:treasury",
			Asset:       "DZD.2",
			Amount:      400,
		}},
	}}); err != nil {
		t.Fatal("seed tx2:", err)
	}
}

func TestLensOverview(t *testing.T) {
	withLensAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		seedLensData(t, l)

		code, out := do(t, r, "GET", "/lenstest/lens/overview", nil)
		if code != http.StatusOK {
			t.Fatalf("overview: expected 200, got %d: %v", code, out)
		}
		data, ok := out["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("overview: expected data object, got %v", out)
		}
		txCount, ok := data["transactions"].(float64)
		if !ok || txCount < 2 {
			t.Fatalf("overview: expected transactions >= 2, got %v", data["transactions"])
		}
		vba, ok := data["volume_by_asset"]
		if !ok {
			t.Fatalf("overview: missing volume_by_asset key, got %v", data)
		}
		arr, ok := vba.([]interface{})
		if !ok || len(arr) == 0 {
			t.Fatalf("overview: expected non-empty volume_by_asset, got %v", vba)
		}
	})
}

func TestLensFlows(t *testing.T) {
	withLensAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		seedLensData(t, l)

		code, out := do(t, r, "GET", "/lenstest/lens/flows", nil)
		if code != http.StatusOK {
			t.Fatalf("flows: expected 200, got %d: %v", code, out)
		}
		rawData, ok := out["data"]
		if !ok {
			t.Fatalf("flows: missing data key, got %v", out)
		}
		arr, ok := rawData.([]interface{})
		if !ok || len(arr) == 0 {
			t.Fatalf("flows: expected non-empty array, got %v", rawData)
		}
		first, ok := arr[0].(map[string]interface{})
		if !ok {
			t.Fatalf("flows: first element is not an object: %v", arr[0])
		}
		for _, key := range []string{"source", "destination", "asset", "time_bucket", "amount", "count"} {
			if _, exists := first[key]; !exists {
				t.Fatalf("flows: first edge missing key %q, got %v", key, first)
			}
		}
	})
}

func TestLensFlowsLimitParam(t *testing.T) {
	withLensAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		seedLensData(t, l)

		// limit=1 — should still return 200 with an array
		code, out := do(t, r, "GET", "/lenstest/lens/flows?limit=1", nil)
		if code != http.StatusOK {
			t.Fatalf("flows limit: expected 200, got %d: %v", code, out)
		}
		arr, ok := out["data"].([]interface{})
		if !ok {
			t.Fatalf("flows limit: expected array data, got %v", out["data"])
		}
		if len(arr) > 1 {
			t.Fatalf("flows limit: expected at most 1 edge, got %d", len(arr))
		}
	})
}

func TestLensRollup(t *testing.T) {
	withLensAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		seedLensData(t, l)

		code, out := do(t, r, "GET", "/lenstest/lens/rollup", nil)
		if code != http.StatusOK {
			t.Fatalf("rollup: expected 200, got %d: %v", code, out)
		}
		data, ok := out["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("rollup: expected data object, got %v", out)
		}
		if _, ok := data["contracts_by_state"]; !ok {
			t.Fatalf("rollup: missing contracts_by_state key, got %v", data)
		}
		if _, ok := data["guard_events_by_action"]; !ok {
			t.Fatalf("rollup: missing guard_events_by_action key, got %v", data)
		}
	})
}

func TestLensTimeSeriesMissingParams(t *testing.T) {
	withLensAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		// No query params → 400
		code, out := do(t, r, "GET", "/lenstest/lens/timeseries", nil)
		if code != http.StatusBadRequest {
			t.Fatalf("timeseries no params: expected 400, got %d: %v", code, out)
		}
	})
}

func TestLensTimeSeriesWithParams(t *testing.T) {
	withLensAPI(t, func(r *gin.Engine, l *ledger.Ledger) {
		seedLensData(t, l)

		code, out := do(t, r, "GET", "/lenstest/lens/timeseries?account=@client:a&asset=DZD.2", nil)
		if code != http.StatusOK {
			t.Fatalf("timeseries: expected 200, got %d: %v", code, out)
		}
		rawData, ok := out["data"]
		if !ok {
			t.Fatalf("timeseries: missing data key, got %v", out)
		}
		if _, ok := rawData.([]interface{}); !ok {
			t.Fatalf("timeseries: expected array, got %T: %v", rawData, rawData)
		}
	})
}
