package guard

import (
	"encoding/json"
	"testing"

	"github.com/amezianechayer/corren/core"
)

func TestScopeMatch(t *testing.T) {
	cases := []struct {
		pattern, account string
		want             bool
	}{
		{"@client:*", "@client:anis", true},
		{"@client:*", "@clientx", false},
		{"@contracts:*", "@contracts:mur1:receivable", true},
		{"@bank:treasury", "@bank:treasury", true},
		{"@bank:treasury", "@bank:income", false},
		{"*", "@anything", true},
	}
	for _, c := range cases {
		if got := scopeMatch(c.pattern, c.account); got != c.want {
			t.Errorf("scopeMatch(%q,%q)=%v want %v", c.pattern, c.account, got, c.want)
		}
	}
}

func tx(postings ...core.Posting) []core.Transaction {
	return []core.Transaction{{Postings: postings}}
}

func netFlows(txs []core.Transaction) map[string]map[string]int64 {
	rf := map[string]map[string]int64{}
	for _, t := range txs {
		for _, p := range t.Postings {
			if rf[p.Source] == nil {
				rf[p.Source] = map[string]int64{}
			}
			rf[p.Source][p.Asset] += p.Amount
			if rf[p.Destination] == nil {
				rf[p.Destination] = map[string]int64{}
			}
			rf[p.Destination][p.Asset] -= p.Amount
		}
	}
	return rf
}

func TestAmountCapPostingBasis(t *testing.T) {
	raw, _ := json.Marshal(AmountCapParams{Scope: "@client:*", Asset: "DZD.2", Max: 1000, Basis: "posting"})
	r := Rule{Kind: KindAmountCap, Params: raw}
	txs := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 1500})
	if matched, _ := evalRule(r, txs, netFlows(txs)); !matched {
		t.Fatal("posting 1500 > cap 1000 from @client:* must match")
	}
	txs2 := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 800})
	if matched, _ := evalRule(r, txs2, netFlows(txs2)); matched {
		t.Fatal("posting 800 <= cap 1000 must not match")
	}
	txs3 := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "EUR.2", Amount: 5000})
	if matched, _ := evalRule(r, txs3, netFlows(txs3)); matched {
		t.Fatal("different asset must not match")
	}
}

func TestAmountCapNetOutflow(t *testing.T) {
	raw, _ := json.Marshal(AmountCapParams{Scope: "@client:anis", Asset: "DZD.2", Max: 1000, Basis: "net_outflow"})
	r := Rule{Kind: KindAmountCap, Params: raw}
	txs := tx(
		core.Posting{Source: "@client:anis", Destination: "@a", Asset: "DZD.2", Amount: 1500},
		core.Posting{Source: "@b", Destination: "@client:anis", Asset: "DZD.2", Amount: 700},
	)
	if matched, _ := evalRule(r, txs, netFlows(txs)); matched {
		t.Fatal("net outflow 800 <= 1000 must not match")
	}
	txs2 := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "DZD.2", Amount: 1500})
	if matched, _ := evalRule(r, txs2, netFlows(txs2)); !matched {
		t.Fatal("net outflow 1500 > 1000 must match")
	}
}

func TestAccountListBlockAndAllow(t *testing.T) {
	block, _ := json.Marshal(AccountListParams{Mode: "block", Side: "destination", Patterns: []string{"@sanctioned:*"}})
	rb := Rule{Kind: KindAccountList, Params: block}
	txs := tx(core.Posting{Source: "@client:anis", Destination: "@sanctioned:x", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(rb, txs, netFlows(txs)); !matched {
		t.Fatal("blocklisted destination must match")
	}
	ok := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(rb, ok, netFlows(ok)); matched {
		t.Fatal("non-blocklisted destination must not match")
	}

	allow, _ := json.Marshal(AccountListParams{Mode: "allow", Side: "destination", Patterns: []string{"@bank:*"}})
	ra := Rule{Kind: KindAccountList, Params: allow}
	bad := tx(core.Posting{Source: "@client:anis", Destination: "@stranger:y", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(ra, bad, netFlows(bad)); !matched {
		t.Fatal("destination outside allowlist must match (deny)")
	}
	good := tx(core.Posting{Source: "@client:anis", Destination: "@bank:treasury", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(ra, good, netFlows(good)); matched {
		t.Fatal("destination in allowlist must not match")
	}
}

func TestAssetRestrict(t *testing.T) {
	only, _ := json.Marshal(AssetRestrictParams{Scope: "@client:*", Mode: "only", Assets: []string{"DZD.2"}})
	r := Rule{Kind: KindAssetRestrict, Params: only}
	bad := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "EUR.2", Amount: 10})
	if matched, _ := evalRule(r, bad, netFlows(bad)); !matched {
		t.Fatal("client transacting a non-allowed asset must match")
	}
	good := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "DZD.2", Amount: 10})
	if matched, _ := evalRule(r, good, netFlows(good)); matched {
		t.Fatal("client transacting the only-allowed asset must not match")
	}

	never, _ := json.Marshal(AssetRestrictParams{Scope: "@client:*", Mode: "never", Assets: []string{"GOLD"}})
	rn := Rule{Kind: KindAssetRestrict, Params: never}
	g := tx(core.Posting{Source: "@client:anis", Destination: "@a", Asset: "GOLD", Amount: 1})
	if matched, _ := evalRule(rn, g, netFlows(g)); !matched {
		t.Fatal("never-allowed asset must match")
	}
}
