package core

import "fmt"

type ShariaViolation struct {
	Constraint string
	Reason     string
}

func (e *ShariaViolation) Error() string {
	return fmt.Sprintf("sharia violation [%s]: %s", e.Constraint, e.Reason)
}

type ShariaConstraint interface {
	Name() string
	Validate(funding Funding, dest Account) error
}
