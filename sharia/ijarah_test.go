package sharia

import "testing"

func validIjarah() IjarahParams {
	return IjarahParams{
		AssetCode: "VHCL1", Cost: Monetary{Asset: "DZD.2", Amount: 10000000},
		Rent: Monetary{Asset: "DZD.2", Amount: 500000}, Client: "@client:anis",
		Supplier: "@supplier:toyota", BankTreasury: "@bank:treasury",
		Periods: 24, FirstDue: "2026-07-01T00:00:00Z", PeriodDays: 30,
	}
}

func TestIjarahValidateOK(t *testing.T) {
	p := validIjarah()
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestIjarahValidateDefaults(t *testing.T) {
	p := validIjarah()
	p.BankTreasury, p.PeriodDays = "", 0
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.BankTreasury != "@bank:treasury" || p.PeriodDays != 30 {
		t.Fatalf("defaults not applied: %+v", p)
	}
}

func TestIjarahValidateRejects(t *testing.T) {
	cases := []struct {
		name, detail string
		mut          func(*IjarahParams)
	}{
		{"zero cost", "cost", func(p *IjarahParams) { p.Cost.Amount = 0 }},
		{"zero rent", "rent", func(p *IjarahParams) { p.Rent.Amount = 0 }},
		{"asset mismatch", "asset", func(p *IjarahParams) { p.Rent.Asset = "EUR.2" }},
		{"asset_code equals money", "asset_code", func(p *IjarahParams) { p.AssetCode = "DZD"; p.Cost.Asset = "DZD"; p.Rent.Asset = "DZD" }},
		{"zero periods", "periods", func(p *IjarahParams) { p.Periods = 0 }},
		{"rent under cost", "exceed cost", func(p *IjarahParams) { p.Rent.Amount = 100; p.Periods = 2 }},
		{"bad client", "client", func(p *IjarahParams) { p.Client = "anis" }},
		{"bad due", "first_due", func(p *IjarahParams) { p.FirstDue = "nope" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := validIjarah()
			c.mut(&p)
			err := p.Validate()
			if err == nil {
				t.Fatal("expected error")
			}
			se, ok := err.(*Error)
			if !ok || se.Code != ErrInvalidParams {
				t.Fatalf("expected ERR_INVALID_PARAMS, got %v", err)
			}
		})
	}
}

func TestBuildIjarahScheduleReference(t *testing.T) {
	items, err := BuildIjarahSchedule(10000000, 500000, 24, "2026-07-01T00:00:00Z", 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 24 {
		t.Fatalf("expected 24, got %d", len(items))
	}
	var rent, depr int64
	for _, it := range items {
		if it.Amount != 500000 {
			t.Fatalf("seq %d rent: expected 500000, got %d", it.Seq, it.Amount)
		}
		if it.Status != StatusPending {
			t.Fatalf("seq %d: expected pending", it.Seq)
		}
		rent += it.Amount
		depr += it.DepreciationPart
	}
	if rent != 12000000 {
		t.Fatalf("total rent: expected 12000000, got %d", rent)
	}
	if depr != 10000000 {
		t.Fatalf("total depreciation: expected 10000000, got %d", depr)
	}
	for i := 0; i < 23; i++ {
		if items[i].DepreciationPart != 416666 {
			t.Fatalf("seq %d depr: expected 416666, got %d", i+1, items[i].DepreciationPart)
		}
	}
	if items[23].DepreciationPart != 416682 {
		t.Fatalf("seq 24 depr: expected 416682, got %d", items[23].DepreciationPart)
	}
}
