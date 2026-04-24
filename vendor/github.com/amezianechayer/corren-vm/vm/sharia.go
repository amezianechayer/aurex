package vm

import (
	"strings"

	"github.com/amezianechayer/corren-vm/core"
)

var ribaAccountPrefixes = []string{
	"@interest:",
	"@riba:",
	"@usury:",
	"@loan_interest:",
}

type NoRibaConstraint struct{}

func (c *NoRibaConstraint) Name() string { return "NO_RIBA" }

func (c *NoRibaConstraint) Validate(funding core.Funding, dest core.Account) error {
	d := strings.ToLower(string(dest))
	for _, prefix := range ribaAccountPrefixes {
		if strings.HasPrefix(d, prefix) {
			return &core.ShariaViolation{
				Constraint: c.Name(),
				Reason:     string(dest) + " is an interest-bearing account (riba is prohibited)",
			}
		}
	}
	return nil
}

type NoGhararConstraint struct{}

func (c *NoGhararConstraint) Name() string { return "NO_GHARAR" }

func (c *NoGhararConstraint) Validate(funding core.Funding, dest core.Account) error {
	if funding.Infinite {
		return nil
	}
	var total uint64
	for _, part := range funding.Parts {
		total += part.Amount
	}
	if total == 0 {
		return &core.ShariaViolation{
			Constraint: c.Name(),
			Reason:     "transfer amount is undefined (zero) — ambiguity in transaction value constitutes gharar",
		}
	}
	return nil
}

var maysirAccountPrefixes = []string{
	"@speculation:",
	"@derivative:",
	"@gambling:",
	"@maysir:",
	"@option:",
	"@futures:",
	"@swap:",
}

type NoMaysirConstraint struct{}

func (c *NoMaysirConstraint) Name() string { return "NO_MAYSIR" }

func (c *NoMaysirConstraint) Validate(funding core.Funding, dest core.Account) error {
	d := strings.ToLower(string(dest))
	for _, prefix := range maysirAccountPrefixes {
		if strings.HasPrefix(d, prefix) {
			return &core.ShariaViolation{
				Constraint: c.Name(),
				Reason:     string(dest) + " is a speculative account (maysir/gambling is prohibited)",
			}
		}
	}
	return nil
}

type AssetBackedConstraint struct {
	AllowedAssets map[string]bool
}

func NewAssetBackedConstraint(assets []string) *AssetBackedConstraint {
	m := make(map[string]bool, len(assets))
	for _, a := range assets {
		m[strings.ToUpper(a)] = true
	}
	return &AssetBackedConstraint{AllowedAssets: m}
}

func (c *AssetBackedConstraint) Name() string { return "ASSET_BACKED" }

func (c *AssetBackedConstraint) Validate(funding core.Funding, dest core.Account) error {
	if len(c.AllowedAssets) == 0 {
		return nil
	}
	fromWorld := funding.Infinite
	if !fromWorld {
		for _, part := range funding.Parts {
			if part.Account == "@world" {
				fromWorld = true
				break
			}
		}
	}
	if !fromWorld {
		return nil
	}
	asset := strings.ToUpper(string(funding.Asset))
	if dot := strings.Index(asset, "."); dot != -1 {
		asset = asset[:dot]
	}
	if !c.AllowedAssets[asset] {
		return &core.ShariaViolation{
			Constraint: c.Name(),
			Reason:     string(funding.Asset) + " is not in the approved asset registry",
		}
	}
	return nil
}
