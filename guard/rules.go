package guard

import (
	"encoding/json"
	"strings"

	"github.com/amezianechayer/corren/core"
)

// scopeMatch: prefix glob. "@client:*" matches "@client:anis"; "*" matches all;
// otherwise exact match. No regex (lexer-safe, YAGNI).
func scopeMatch(pattern, account string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(account, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == account
}

func anyMatch(patterns []string, account string) bool {
	for _, p := range patterns {
		if scopeMatch(p, account) {
			return true
		}
	}
	return false
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// evalRule reports whether the rule MATCHES (i.e. should fire) against the
// proposed transactions. A match means: deny (if action=deny) or record a
// monitor event (if action=monitor). The string is a short detail for the
// event payload. Pure: reads only txs + netFlows.
func evalRule(r Rule, txs []core.Transaction, nf map[string]map[string]int64) (bool, string) {
	switch r.Kind {
	case KindAmountCap:
		var p AmountCapParams
		if json.Unmarshal(r.Params, &p) != nil {
			return false, ""
		}
		if p.Basis == "net_outflow" {
			for acct, assets := range nf {
				if !scopeMatch(p.Scope, acct) {
					continue
				}
				if out := assets[p.Asset]; out > p.Max {
					return true, "net_outflow " + acct
				}
			}
			return false, ""
		}
		// default: posting basis
		for _, t := range txs {
			for _, post := range t.Postings {
				if post.Asset == p.Asset && scopeMatch(p.Scope, post.Source) && post.Amount > p.Max {
					return true, "posting " + post.Source
				}
			}
		}
		return false, ""

	case KindAccountList:
		var p AccountListParams
		if json.Unmarshal(r.Params, &p) != nil {
			return false, ""
		}
		sideHit := func(post core.Posting) bool {
			switch p.Side {
			case "source":
				return anyMatch(p.Patterns, post.Source)
			case "destination":
				return anyMatch(p.Patterns, post.Destination)
			default: // either
				return anyMatch(p.Patterns, post.Source) || anyMatch(p.Patterns, post.Destination)
			}
		}
		for _, t := range txs {
			for _, post := range t.Postings {
				hit := sideHit(post)
				if p.Mode == "block" && hit {
					return true, "blocked " + post.Source + "->" + post.Destination
				}
				if p.Mode == "allow" && !hit {
					return true, "not allowlisted " + post.Source + "->" + post.Destination
				}
			}
		}
		return false, ""

	case KindAssetRestrict:
		var p AssetRestrictParams
		if json.Unmarshal(r.Params, &p) != nil {
			return false, ""
		}
		for _, t := range txs {
			for _, post := range t.Postings {
				inScope := scopeMatch(p.Scope, post.Source) || scopeMatch(p.Scope, post.Destination)
				if !inScope {
					continue
				}
				listed := contains(p.Assets, post.Asset)
				if p.Mode == "only" && !listed {
					return true, "asset " + post.Asset + " not allowed"
				}
				if p.Mode == "never" && listed {
					return true, "asset " + post.Asset + " forbidden"
				}
			}
		}
		return false, ""
	}
	return false, ""
}
