package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/amezianechayer/corren/payments"
	"github.com/huandu/go-sqlbuilder"
)

var paymentCols = []string{"id", "psp", "direction", "wallet_id", "asset", "amount", "status", "reference", "external_id", "created_at", "updated_at"}

func (s *SQLiteStore) SavePayment(p payments.Payment) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("payments")
	ib.Cols(paymentCols...)
	ib.Values(p.ID, p.PSP, p.Direction, p.WalletID, p.Asset, p.Amount, p.Status, p.Reference, p.ExternalID, p.CreatedAt, p.UpdatedAt)
	return s.exec(ib)
}

func (s *SQLiteStore) UpdatePayment(p payments.Payment) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("payments")
	ub.Set(
		ub.Assign("status", p.Status),
		ub.Assign("external_id", p.ExternalID),
		ub.Assign("updated_at", p.UpdatedAt),
	)
	ub.Where(ub.Equal("id", p.ID))
	return s.exec(ub)
}

func (s *SQLiteStore) GetPayment(id string) (payments.Payment, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(paymentCols...)
	sb.From("payments")
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)

	var p payments.Payment
	err := s.db.QueryRow(sqlq, args...).Scan(
		&p.ID, &p.PSP, &p.Direction, &p.WalletID, &p.Asset, &p.Amount, &p.Status, &p.Reference, &p.ExternalID, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return payments.Payment{}, fmt.Errorf("payment not found: %s", id)
	}
	return p, err
}

func (s *SQLiteStore) ListPayments(limit, offset int) ([]payments.Payment, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(paymentCols...)
	sb.From("payments")
	sb.OrderBy("created_at").Desc()
	sb.Limit(limit).Offset(offset)
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)

	rows, err := s.db.Query(sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []payments.Payment{}
	for rows.Next() {
		var p payments.Payment
		if err := rows.Scan(
			&p.ID, &p.PSP, &p.Direction, &p.WalletID, &p.Asset, &p.Amount, &p.Status, &p.Reference, &p.ExternalID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
