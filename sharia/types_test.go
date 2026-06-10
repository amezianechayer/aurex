package sharia

import (
	"strings"
	"testing"
)

func validParams() MurabahaParams {
	return MurabahaParams{
		AssetCode:    "VHCL42A",
		Cost:         Monetary{Asset: "SAR2", Amount: 10000000},
		Markup:       Monetary{Asset: "SAR2", Amount: 1000000},
		Client:       "@client:ameziane",
		Supplier:     "@supplier:toyota",
		BankTreasury: "@bank:treasury",
		Installments: 24,
		FirstDue:     "2026-07-01T00:00:00Z",
		PeriodDays:   30,
	}
}

func TestValidateParamsOK(t *testing.T) {
	p := validParams()
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid params, got %v", err)
	}
}

func TestValidateParamsDefaults(t *testing.T) {
	p := validParams()
	p.BankTreasury = ""
	p.PeriodDays = 0
	if err := p.Validate(); err != nil {
		t.Fatalf("expected valid params with defaults, got %v", err)
	}
	if p.BankTreasury != "@bank:treasury" {
		t.Errorf("expected default treasury, got %q", p.BankTreasury)
	}
	if p.PeriodDays != 30 {
		t.Errorf("expected default period 30, got %d", p.PeriodDays)
	}
}

func TestValidateParamsRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*MurabahaParams)
		detail string // substring expected in error message
	}{
		{"zero cost", func(p *MurabahaParams) { p.Cost.Amount = 0 }, "cost"},
		{"negative markup", func(p *MurabahaParams) { p.Markup.Amount = -1 }, "markup"},
		{"asset mismatch", func(p *MurabahaParams) { p.Markup.Asset = "USD2" }, "asset"},
		{"bad monetary asset", func(p *MurabahaParams) { p.Cost.Asset = "sar2"; p.Markup.Asset = "sar2" }, "asset"},
		{"bad asset code", func(p *MurabahaParams) { p.AssetCode = "vhcl" }, "asset_code"},
		{"asset code equals monetary", func(p *MurabahaParams) { p.AssetCode = "SAR2" }, "asset_code"},
		{"zero installments", func(p *MurabahaParams) { p.Installments = 0 }, "installments"},
		{"bad first due", func(p *MurabahaParams) { p.FirstDue = "tomorrow" }, "first_due"},
		{"bad client account", func(p *MurabahaParams) { p.Client = "client" }, "client"},
		{"bad supplier account", func(p *MurabahaParams) { p.Supplier = "@Supplier" }, "supplier"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := validParams()
			c.mutate(&p)
			err := p.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			se, ok := err.(*Error)
			if !ok {
				t.Fatalf("expected *sharia.Error, got %T", err)
			}
			if se.Code != ErrInvalidParams {
				t.Fatalf("expected %s, got %s", ErrInvalidParams, se.Code)
			}
			if !strings.Contains(se.Message, c.detail) {
				t.Errorf("expected message to mention %q, got %q", c.detail, se.Message)
			}
		})
	}
}

func TestContractIDValidation(t *testing.T) {
	valid := []string{"mur20260610a7x4", "abcd", "m_42", "a123456789"}
	invalid := []string{"", "ab", "Mur42", "mur-42", "42mur", "_x42", strings.Repeat("a", 42)}

	for _, id := range valid {
		if !ContractIDIsValid(id) {
			t.Errorf("expected %q to be valid", id)
		}
	}
	for _, id := range invalid {
		if ContractIDIsValid(id) {
			t.Errorf("expected %q to be invalid", id)
		}
	}
}

func TestGenerateContractID(t *testing.T) {
	id1 := GenerateContractID()
	id2 := GenerateContractID()
	if !ContractIDIsValid(id1) || !ContractIDIsValid(id2) {
		t.Fatalf("generated ids must be valid: %q %q", id1, id2)
	}
	if !strings.HasPrefix(id1, "mur") {
		t.Errorf("expected mur prefix, got %q", id1)
	}
	if id1 == id2 {
		t.Errorf("expected distinct ids, got %q twice", id1)
	}
}

func TestErrorHTTPStatus(t *testing.T) {
	cases := map[string]int{
		ErrInvalidParams:     400,
		ErrNotFound:          404,
		ErrInvalidTransition: 409,
		ErrDuplicate:         409,
		ErrPrecondition:      422,
		ErrShariaViolation:   422,
	}
	for code, status := range cases {
		e := &Error{Code: code}
		if s := e.HTTPStatus(); s != status {
			t.Errorf("%s: expected %d, got %d", code, status, s)
		}
	}
}
