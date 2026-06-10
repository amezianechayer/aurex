package auth

import "testing"

func TestRoleRank(t *testing.T) {
	if !RoleAtLeast(RoleAdmin, RoleOperator) || !RoleAtLeast(RoleOperator, RoleReadonly) {
		t.Fatal("admin>=operator>=readonly expected")
	}
	if RoleAtLeast(RoleReadonly, RoleOperator) {
		t.Fatal("readonly must not satisfy operator")
	}
	if RoleAtLeast("bogus", RoleReadonly) {
		t.Fatal("unknown role must rank below everything")
	}
	if !RoleIsValid(RoleAdmin) || RoleIsValid("bogus") {
		t.Fatal("RoleIsValid broken")
	}
}

func TestIdentityAllowsLedger(t *testing.T) {
	all := Identity{Role: RoleOperator, Ledgers: []string{"*"}}
	one := Identity{Role: RoleOperator, Ledgers: []string{"demo"}}
	if !all.AllowsLedger("anything") || !one.AllowsLedger("demo") || one.AllowsLedger("prod") {
		t.Fatal("ledger scope broken")
	}
}
