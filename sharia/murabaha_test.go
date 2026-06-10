package sharia

import (
	"testing"

	"github.com/amezianechayer/corren/core"
)

const cid = "mur20260610a7x4"

func assertPosting(t *testing.T, p core.Posting, source, destination, asset string, amount int64) {
	t.Helper()
	if p.Source != source || p.Destination != destination || p.Asset != asset || p.Amount != amount {
		t.Fatalf("posting mismatch: expected {%s -> %s [%s %d]}, got {%s -> %s [%s %d]}",
			source, destination, asset, amount, p.Source, p.Destination, p.Asset, p.Amount)
	}
}

func TestFSMAllowedTransitions(t *testing.T) {
	allowed := []struct{ state, name string }{
		{StatePromise, "acquire"},
		{StatePromise, "cancel"},
		{StateAcquired, "sell"},
		{StateAcquired, "cancel"},
		{StateSold, "pay_installment"},
		{StateSold, "early_settle"},
		{StateSold, "late_penalty"},
	}
	denied := []struct{ state, name string }{
		{StatePromise, "sell"},
		{StatePromise, "pay_installment"},
		{StateAcquired, "pay_installment"},
		{StateAcquired, "early_settle"},
		{StateSold, "cancel"}, // rule C-2: a concluded sale cannot be cancelled
		{StateSold, "acquire"},
		{StateSettled, "pay_installment"},
		{StateCancelled, "acquire"},
		{StateSold, "unknown"},
	}

	for _, c := range allowed {
		if !TransitionAllowed(c.state, c.name) {
			t.Errorf("expected %s allowed from %s", c.name, c.state)
		}
	}
	for _, c := range denied {
		if TransitionAllowed(c.state, c.name) {
			t.Errorf("expected %s denied from %s", c.name, c.state)
		}
	}
}

func TestAcquirePostings(t *testing.T) {
	p := validParams()
	postings := AcquirePostings(cid, p)
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings, got %d", len(postings))
	}
	assertPosting(t, postings[0], "@bank:treasury", "@supplier:toyota", "SAR2", 10000000)
	assertPosting(t, postings[1], "@world", "@contracts:"+cid+":inventory", "VHCL42A", 1)
}

func TestSellPostings(t *testing.T) {
	p := validParams()
	postings := SellPostings(cid, p)
	if len(postings) != 3 {
		t.Fatalf("expected 3 postings, got %d", len(postings))
	}
	assertPosting(t, postings[0], "@contracts:"+cid+":inventory", "@client:ameziane:assets", "VHCL42A", 1)
	assertPosting(t, postings[1], "@contracts:"+cid+":counterpart", "@contracts:"+cid+":receivable", "SAR2", 11000000)
	assertPosting(t, postings[2], "@contracts:"+cid+":counterpart", "@contracts:"+cid+":deferred", "SAR2", 1000000)
}

func TestPayInstallmentPostings(t *testing.T) {
	p := validParams()
	inst := Installment{Seq: 1, Amount: 458333, PrincipalPart: 416667, ProfitPart: 41666}
	postings := PayInstallmentPostings(cid, p, inst)
	if len(postings) != 3 {
		t.Fatalf("expected 3 postings, got %d", len(postings))
	}
	assertPosting(t, postings[0], "@client:ameziane", "@bank:treasury", "SAR2", 458333)
	assertPosting(t, postings[1], "@contracts:"+cid+":receivable", "@contracts:"+cid+":counterpart", "SAR2", 458333)
	assertPosting(t, postings[2], "@contracts:"+cid+":deferred", "@bank:income:murabaha", "SAR2", 41666)
}

func TestPayInstallmentPostingsZeroProfit(t *testing.T) {
	p := validParams()
	inst := Installment{Seq: 1, Amount: 100, PrincipalPart: 100, ProfitPart: 0}
	postings := PayInstallmentPostings(cid, p, inst)
	if len(postings) != 2 {
		t.Fatalf("expected 2 postings when profit part is zero, got %d", len(postings))
	}
}

func TestEarlySettlePostingsFullRebate(t *testing.T) {
	p := validParams()
	// 9 unpaid installments left: rest_total, rest_profit; full rebate
	postings := EarlySettlePostings(cid, p, 4125006, 375006, 375006)
	if len(postings) != 3 {
		t.Fatalf("expected 3 postings (P3 omitted), got %d", len(postings))
	}
	assertPosting(t, postings[0], "@client:ameziane", "@bank:treasury", "SAR2", 4125006-375006)
	assertPosting(t, postings[1], "@contracts:"+cid+":receivable", "@contracts:"+cid+":counterpart", "SAR2", 4125006)
	// P3 omitted (rest_profit - rebate == 0), P4 cancels the rebated profit
	assertPosting(t, postings[2], "@contracts:"+cid+":deferred", "@contracts:"+cid+":counterpart", "SAR2", 375006)
}

func TestEarlySettlePostingsNoRebate(t *testing.T) {
	p := validParams()
	postings := EarlySettlePostings(cid, p, 4125006, 375006, 0)
	if len(postings) != 3 {
		t.Fatalf("expected 3 postings (P4 omitted), got %d", len(postings))
	}
	assertPosting(t, postings[0], "@client:ameziane", "@bank:treasury", "SAR2", 4125006)
	assertPosting(t, postings[1], "@contracts:"+cid+":receivable", "@contracts:"+cid+":counterpart", "SAR2", 4125006)
	assertPosting(t, postings[2], "@contracts:"+cid+":deferred", "@bank:income:murabaha", "SAR2", 375006)
}

func TestEarlySettlePostingsPartialRebate(t *testing.T) {
	p := validParams()
	postings := EarlySettlePostings(cid, p, 1000, 300, 100)
	if len(postings) != 4 {
		t.Fatalf("expected 4 postings, got %d", len(postings))
	}
	assertPosting(t, postings[0], "@client:ameziane", "@bank:treasury", "SAR2", 900)
	assertPosting(t, postings[1], "@contracts:"+cid+":receivable", "@contracts:"+cid+":counterpart", "SAR2", 1000)
	assertPosting(t, postings[2], "@contracts:"+cid+":deferred", "@bank:income:murabaha", "SAR2", 200)
	assertPosting(t, postings[3], "@contracts:"+cid+":deferred", "@contracts:"+cid+":counterpart", "SAR2", 100)
}

func TestPenaltyPostings(t *testing.T) {
	p := validParams()
	postings, err := PenaltyPostings(cid, p, 20000, "@charity:pool")
	if err != nil {
		t.Fatal(err)
	}
	if len(postings) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(postings))
	}
	assertPosting(t, postings[0], "@client:ameziane", "@charity:pool", "SAR2", 20000)
}

func TestPenaltyPostingsRejectsNonCharityDestination(t *testing.T) {
	p := validParams()
	for _, dest := range []string{"@bank:income:murabaha", "@bank:treasury", "@client:other", "charity:pool"} {
		_, err := PenaltyPostings(cid, p, 20000, dest)
		if err == nil {
			t.Fatalf("expected I-3 violation for destination %q", dest)
		}
		se, ok := err.(*Error)
		if !ok || se.Code != ErrShariaViolation {
			t.Fatalf("expected ERR_SHARIA_VIOLATION, got %v", err)
		}
		if se.StandardRef != RefSS3 {
			t.Fatalf("expected standard_ref %s, got %s", RefSS3, se.StandardRef)
		}
	}
}

func TestCancelPostingsFromAcquired(t *testing.T) {
	p := validParams()
	postings := CancelPostings(cid, p, StateAcquired)
	if len(postings) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(postings))
	}
	assertPosting(t, postings[0], "@contracts:"+cid+":inventory", "@bank:inventory:unsold", "VHCL42A", 1)
}

func TestCancelPostingsFromPromise(t *testing.T) {
	p := validParams()
	postings := CancelPostings(cid, p, StatePromise)
	if len(postings) != 0 {
		t.Fatalf("expected no postings from PROMISE, got %d", len(postings))
	}
}
