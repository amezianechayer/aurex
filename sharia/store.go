package sharia

// ShariaStore is the persistence port for contracts, schedules and the
// audit chain. Implemented by storage/sqlite and storage/postgres.
type ShariaStore interface {
	SaveContract(c Contract) error
	UpdateContractState(id, state, updatedAt string) error
	GetContract(id string) (Contract, error)
	ListContracts(limit, offset int) ([]Contract, error)

	SaveSchedule(contractID string, items []Installment) error
	GetSchedule(contractID string) ([]Installment, error)
	MarkInstallment(contractID string, seq int, status string, txID int64, paidAt string) error
	FindDue(nowRFC3339 string, graceDays int) ([]DueItem, error)

	AppendAudit(e AuditEvent) error
	GetAudit(contractID string) ([]AuditEvent, error)
	LastAuditHash(contractID string) (seq int, hash string, err error)
}
