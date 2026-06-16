package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/amezianechayer/corren/wallets"
	"github.com/huandu/go-sqlbuilder"
	"github.com/spf13/viper"
)

func (s *SQLiteStore) exec(b sqlbuilder.Builder) error {
	sqlq, args := b.BuildWithFlavor(sqlbuilder.SQLite)
	if viper.GetBool("debug") {
		fmt.Println(sqlq, args)
	}
	_, err := s.db.Exec(sqlq, args...)
	return err
}

func (s *SQLiteStore) SaveWallet(w wallets.Wallet) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("wallets")
	ib.Cols("id", "owner", "asset", "created_at", "updated_at")
	ib.Values(w.ID, w.Owner, w.Asset, w.CreatedAt, w.UpdatedAt)
	return s.exec(ib)
}

func (s *SQLiteStore) GetWallet(id string) (wallets.Wallet, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "owner", "asset", "created_at", "updated_at")
	sb.From("wallets")
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)

	var w wallets.Wallet
	err := s.db.QueryRow(sqlq, args...).Scan(&w.ID, &w.Owner, &w.Asset, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return wallets.Wallet{}, fmt.Errorf("wallet not found: %s", id)
	}
	return w, err
}

func (s *SQLiteStore) ListWallets(limit, offset int) ([]wallets.Wallet, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "owner", "asset", "created_at", "updated_at")
	sb.From("wallets")
	sb.OrderBy("created_at").Desc()
	sb.Limit(limit).Offset(offset)
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)

	rows, err := s.db.Query(sqlq, args...)
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

func (s *SQLiteStore) SaveHold(h wallets.Hold) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("wallet_holds")
	ib.Cols("id", "wallet_id", "asset", "amount", "status", "reason", "expires_at", "created_at", "updated_at")
	ib.Values(h.ID, h.WalletID, h.Asset, h.Amount, h.Status, h.Reason, h.ExpiresAt, h.CreatedAt, h.UpdatedAt)
	return s.exec(ib)
}

func (s *SQLiteStore) UpdateHold(h wallets.Hold) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("wallet_holds")
	ub.Set(
		ub.Assign("status", h.Status),
		ub.Assign("updated_at", h.UpdatedAt),
	)
	ub.Where(ub.Equal("id", h.ID))
	return s.exec(ub)
}

func (s *SQLiteStore) GetHold(id string) (wallets.Hold, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "wallet_id", "asset", "amount", "status", "reason", "expires_at", "created_at", "updated_at")
	sb.From("wallet_holds")
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)

	var h wallets.Hold
	err := s.db.QueryRow(sqlq, args...).Scan(
		&h.ID, &h.WalletID, &h.Asset, &h.Amount, &h.Status, &h.Reason, &h.ExpiresAt, &h.CreatedAt, &h.UpdatedAt)
	if err == sql.ErrNoRows {
		return wallets.Hold{}, fmt.Errorf("hold not found: %s", id)
	}
	return h, err
}

func (s *SQLiteStore) ListHolds(walletID string) ([]wallets.Hold, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "wallet_id", "asset", "amount", "status", "reason", "expires_at", "created_at", "updated_at")
	sb.From("wallet_holds")
	sb.Where(sb.Equal("wallet_id", walletID))
	sb.OrderBy("created_at").Desc()
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)

	rows, err := s.db.Query(sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []wallets.Hold{}
	for rows.Next() {
		var h wallets.Hold
		if err := rows.Scan(
			&h.ID, &h.WalletID, &h.Asset, &h.Amount, &h.Status, &h.Reason, &h.ExpiresAt, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}
