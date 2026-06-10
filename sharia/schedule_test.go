package sharia

import (
	"testing"
	"time"
)

func sum(parts []int64) int64 {
	var t int64
	for _, p := range parts {
		t += p
	}
	return t
}

func TestSplitEven(t *testing.T) {
	cases := []struct {
		name  string
		total int64
		n     int
		first int64
		last  int64
	}{
		{"reference total", 11000000, 24, 458333, 458341},
		{"reference markup", 1000000, 24, 41666, 41682},
		{"n equals 1", 12345, 1, 12345, 12345},
		{"zero total", 0, 4, 0, 0},
		{"non divisible", 100, 3, 33, 34},
		{"total smaller than n", 2, 5, 0, 2},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			parts := SplitEven(c.total, c.n)
			if len(parts) != c.n {
				t.Fatalf("expected %d parts, got %d", c.n, len(parts))
			}
			if s := sum(parts); s != c.total {
				t.Fatalf("parts sum to %d, expected %d", s, c.total)
			}
			if parts[0] != c.first {
				t.Errorf("first part: expected %d, got %d", c.first, parts[0])
			}
			if parts[len(parts)-1] != c.last {
				t.Errorf("last part: expected %d, got %d", c.last, parts[len(parts)-1])
			}
		})
	}
}

func TestBuildScheduleReference(t *testing.T) {
	// Spec reference: cost=10_000_000, markup=1_000_000, n=24
	items, err := BuildSchedule(10000000, 1000000, 24, "2026-07-01T00:00:00Z", 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 24 {
		t.Fatalf("expected 24 installments, got %d", len(items))
	}

	var amounts, profits, principals int64
	for i, it := range items {
		if it.Seq != i+1 {
			t.Errorf("seq %d: expected %d", it.Seq, i+1)
		}
		if it.Status != StatusPending {
			t.Errorf("seq %d: expected status pending, got %s", it.Seq, it.Status)
		}
		if it.Amount != it.PrincipalPart+it.ProfitPart {
			t.Errorf("seq %d: amount %d != principal %d + profit %d",
				it.Seq, it.Amount, it.PrincipalPart, it.ProfitPart)
		}
		amounts += it.Amount
		profits += it.ProfitPart
		principals += it.PrincipalPart
	}

	if amounts != 11000000 {
		t.Errorf("sum of amounts: expected 11000000, got %d", amounts)
	}
	if profits != 1000000 {
		t.Errorf("sum of profits: expected 1000000, got %d", profits)
	}
	if principals != 10000000 {
		t.Errorf("sum of principals: expected 10000000, got %d", principals)
	}

	// Exact reference values from spec §6
	for i := 0; i < 23; i++ {
		if items[i].Amount != 458333 || items[i].ProfitPart != 41666 || items[i].PrincipalPart != 416667 {
			t.Fatalf("installment %d: expected 458333/41666/416667, got %d/%d/%d",
				i+1, items[i].Amount, items[i].ProfitPart, items[i].PrincipalPart)
		}
	}
	last := items[23]
	if last.Amount != 458341 || last.ProfitPart != 41682 || last.PrincipalPart != 416659 {
		t.Fatalf("installment 24: expected 458341/41682/416659, got %d/%d/%d",
			last.Amount, last.ProfitPart, last.PrincipalPart)
	}

	// Dates: due[i] = FirstDue + (i-1)*PeriodDays, UTC
	first, _ := time.Parse(time.RFC3339, "2026-07-01T00:00:00Z")
	for i, it := range items {
		want := first.AddDate(0, 0, i*30).Format(time.RFC3339)
		if it.DueDate != want {
			t.Fatalf("installment %d: expected due %s, got %s", i+1, want, it.DueDate)
		}
	}
}

func TestBuildScheduleMarkupZero(t *testing.T) {
	items, err := BuildSchedule(1000, 0, 3, "2026-07-01T00:00:00Z", 30)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		if it.ProfitPart != 0 {
			t.Fatalf("expected zero profit, got %d", it.ProfitPart)
		}
	}
	if s := items[0].Amount + items[1].Amount + items[2].Amount; s != 1000 {
		t.Fatalf("expected total 1000, got %d", s)
	}
}

func TestBuildScheduleInvalidDate(t *testing.T) {
	_, err := BuildSchedule(1000, 100, 3, "not-a-date", 30)
	if err == nil {
		t.Fatal("expected error for invalid first due date")
	}
}
