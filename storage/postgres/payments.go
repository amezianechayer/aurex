package postgres

import (
	"context"
	"fmt"

	"github.com/amezianechayer/corren/payments"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v4"
)

var paymentCols = []string{"id", "psp", "direction", "wallet_id", "asset", "amount", "status", "reference", "external_id", "created_at", "updated_at"}

func (s *PGStore) SavePayment(p payments.Payment) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("payments"))
	ib.Cols(paymentCols...)
	ib.Values(p.ID, p.PSP, p.Direction, p.WalletID, p.Asset, p.Amount, p.Status, p.Reference, p.ExternalID, p.CreatedAt, p.UpdatedAt)
	return s.execWallet(ib)
}

func (s *PGStore) UpdatePayment(p payments.Payment) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("payments"))
	ub.Set(
		ub.Assign("status", p.Status),
		ub.Assign("external_id", p.ExternalID),
		ub.Assign("updated_at", p.UpdatedAt),
	)
	ub.Where(ub.Equal("id", p.ID))
	return s.execWallet(ub)
}

func (s *PGStore) GetPayment(id string) (payments.Payment, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(paymentCols...)
	sb.From(s.table("payments"))
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	var p payments.Payment
	err := s.Conn().QueryRow(context.Background(), sqlq, args...).Scan(
		&p.ID, &p.PSP, &p.Direction, &p.WalletID, &p.Asset, &p.Amount, &p.Status, &p.Reference, &p.ExternalID, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return payments.Payment{}, fmt.Errorf("payment not found: %s", id)
	}
	return p, err
}

func (s *PGStore) ListPayments(limit, offset int) ([]payments.Payment, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(paymentCols...)
	sb.From(s.table("payments"))
	sb.OrderBy("created_at").Desc()
	sb.Limit(limit).Offset(offset)
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
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
