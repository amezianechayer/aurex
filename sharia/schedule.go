package sharia

import (
	"fmt"
	"time"
)

// SplitEven splits total into n integer parts:
// parts 1..n-1 = total/n (integer division), part n carries the remainder.
func SplitEven(total int64, n int) []int64 {
	parts := make([]int64, n)
	base := total / int64(n)
	for i := 0; i < n-1; i++ {
		parts[i] = base
	}
	parts[n-1] = total - int64(n-1)*base
	return parts
}

// BuildSchedule builds the installment plan for cost+markup over n periods.
// All arithmetic is int64 minor units (invariant I-6): no floats, ever.
func BuildSchedule(cost, markup int64, n int, firstDue string, periodDays int) ([]Installment, error) {
	first, err := time.Parse(time.RFC3339, firstDue)
	if err != nil {
		return nil, fmt.Errorf("invalid first_due date: %w", err)
	}

	amounts := SplitEven(cost+markup, n)
	profits := SplitEven(markup, n)

	items := make([]Installment, n)
	for i := 0; i < n; i++ {
		items[i] = Installment{
			Seq:           i + 1,
			DueDate:       first.UTC().AddDate(0, 0, i*periodDays).Format(time.RFC3339),
			Amount:        amounts[i],
			PrincipalPart: amounts[i] - profits[i],
			ProfitPart:    profits[i],
			Status:        StatusPending,
			PaidTxID:      -1,
		}
	}
	return items, nil
}
