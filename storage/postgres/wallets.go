package postgres

import (
	"context"
	"fmt"

	"github.com/amezianechayer/corren/wallets"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

func (s *PGStore) execWallet(b sqlbuilder.Builder) error {
	sqlq, args := b.BuildWithFlavor(sqlbuilder.PostgreSQL)
	if viper.GetBool("debug") {
		fmt.Println(sqlq, args)
	}
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) SaveWallet(w wallets.Wallet) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("wallets"))
	ib.Cols("id", "owner", "asset", "created_at", "updated_at")
	ib.Values(w.ID, w.Owner, w.Asset, w.CreatedAt, w.UpdatedAt)
	return s.execWallet(ib)
}

func (s *PGStore) GetWallet(id string) (wallets.Wallet, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "owner", "asset", "created_at", "updated_at")
	sb.From(s.table("wallets"))
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	var w wallets.Wallet
	err := s.Conn().QueryRow(context.Background(), sqlq, args...).Scan(
		&w.ID, &w.Owner, &w.Asset, &w.CreatedAt, &w.UpdatedAt)
	if err == pgx.ErrNoRows {
		return wallets.Wallet{}, fmt.Errorf("wallet not found: %s", id)
	}
	return w, err
}

func (s *PGStore) ListWallets(limit, offset int) ([]wallets.Wallet, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "owner", "asset", "created_at", "updated_at")
	sb.From(s.table("wallets"))
	sb.OrderBy("created_at").Desc()
	sb.Limit(limit).Offset(offset)
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []wallets.Wallet{}
	for rows.Next() {
		var w wallets.Wallet
		if err := rows.Scan(&w.ID, &w.Owner, &w.Asset, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func (s *PGStore) SaveHold(h wallets.Hold) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("wallet_holds"))
	ib.Cols("id", "wallet_id", "asset", "amount", "status", "description", "created_at", "updated_at")
	ib.Values(h.ID, h.WalletID, h.Asset, h.Amount, h.Status, h.Description, h.CreatedAt, h.UpdatedAt)
	return s.execWallet(ib)
}

func (s *PGStore) UpdateHold(h wallets.Hold) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("wallet_holds"))
	ub.Set(
		ub.Assign("status", h.Status),
		ub.Assign("updated_at", h.UpdatedAt),
	)
	ub.Where(ub.Equal("id", h.ID))
	return s.execWallet(ub)
}

func (s *PGStore) GetHold(id string) (wallets.Hold, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "wallet_id", "asset", "amount", "status", "description", "created_at", "updated_at")
	sb.From(s.table("wallet_holds"))
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	var h wallets.Hold
	err := s.Conn().QueryRow(context.Background(), sqlq, args...).Scan(
		&h.ID, &h.WalletID, &h.Asset, &h.Amount, &h.Status, &h.Description, &h.CreatedAt, &h.UpdatedAt)
	if err == pgx.ErrNoRows {
		return wallets.Hold{}, fmt.Errorf("hold not found: %s", id)
	}
	return h, err
}

func (s *PGStore) ListHolds(walletID string) ([]wallets.Hold, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "wallet_id", "asset", "amount", "status", "description", "created_at", "updated_at")
	sb.From(s.table("wallet_holds"))
	sb.Where(sb.Equal("wallet_id", walletID))
	sb.OrderBy("created_at").Desc()
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []wallets.Hold{}
	for rows.Next() {
		var h wallets.Hold
		if err := rows.Scan(
			&h.ID, &h.WalletID, &h.Asset, &h.Amount, &h.Status, &h.Description, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}
