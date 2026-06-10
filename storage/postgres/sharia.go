package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amezianechayer/corren/sharia"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

func (s *PGStore) SaveContract(c sharia.Contract) error {
	params, err := json.Marshal(c.Params)
	if err != nil {
		return err
	}

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("sharia_contracts"))
	ib.Cols("id", "type", "state", "params", "template_version", "created_at", "updated_at")
	ib.Values(c.ID, c.Type, c.State, string(params), c.TemplateVersion, c.CreatedAt, c.UpdatedAt)

	sqlq, args := ib.BuildWithFlavor(sqlbuilder.PostgreSQL)
	if viper.GetBool("debug") {
		fmt.Println(sqlq, args)
	}

	_, err = s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) UpdateContractState(id, state, updatedAt string) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("sharia_contracts"))
	ub.Set(
		ub.Assign("state", state),
		ub.Assign("updated_at", updatedAt),
	)
	ub.Where(ub.Equal("id", id))

	sqlq, args := ub.BuildWithFlavor(sqlbuilder.PostgreSQL)
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) GetContract(id string) (sharia.Contract, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "type", "state", "params", "template_version", "created_at", "updated_at")
	sb.From(s.table("sharia_contracts"))
	sb.Where(sb.Equal("id", id))

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	var c sharia.Contract
	var params string
	err := s.Conn().QueryRow(context.Background(), sqlq, args...).Scan(
		&c.ID, &c.Type, &c.State, &params, &c.TemplateVersion, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return c, err
	}

	err = json.Unmarshal([]byte(params), &c.Params)
	return c, err
}

func (s *PGStore) ListContracts(limit, offset int) ([]sharia.Contract, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "type", "state", "params", "template_version", "created_at", "updated_at")
	sb.From(s.table("sharia_contracts"))
	sb.OrderBy("created_at").Desc()
	sb.Limit(limit)
	sb.Offset(offset)

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contracts := []sharia.Contract{}
	for rows.Next() {
		var c sharia.Contract
		var params string
		if err := rows.Scan(&c.ID, &c.Type, &c.State, &params, &c.TemplateVersion, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(params), &c.Params); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}

func (s *PGStore) SaveSchedule(contractID string, items []sharia.Installment) error {
	ctx := context.Background()
	tx, err := s.Conn().Begin(ctx)
	if err != nil {
		return err
	}

	for _, it := range items {
		ib := sqlbuilder.NewInsertBuilder()
		ib.InsertInto(s.table("sharia_schedule"))
		ib.Cols("contract_id", "seq", "due_date", "amount", "principal_part", "profit_part", "status", "paid_tx_id", "paid_at")
		ib.Values(contractID, it.Seq, it.DueDate, it.Amount, it.PrincipalPart, it.ProfitPart, it.Status, it.PaidTxID, it.PaidAt)

		sqlq, args := ib.BuildWithFlavor(sqlbuilder.PostgreSQL)
		if _, err := tx.Exec(ctx, sqlq, args...); err != nil {
			tx.Rollback(ctx)
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *PGStore) GetSchedule(contractID string) ([]sharia.Installment, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("seq", "due_date", "amount", "principal_part", "profit_part", "status", "paid_tx_id", "paid_at")
	sb.From(s.table("sharia_schedule"))
	sb.Where(sb.Equal("contract_id", contractID))
	sb.OrderBy("seq").Asc()

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []sharia.Installment{}
	for rows.Next() {
		var it sharia.Installment
		if err := rows.Scan(&it.Seq, &it.DueDate, &it.Amount, &it.PrincipalPart, &it.ProfitPart, &it.Status, &it.PaidTxID, &it.PaidAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *PGStore) MarkInstallment(contractID string, seq int, status string, txID int64, paidAt string) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("sharia_schedule"))
	ub.Set(
		ub.Assign("status", status),
		ub.Assign("paid_tx_id", txID),
		ub.Assign("paid_at", paidAt),
	)
	ub.Where(ub.And(
		ub.Equal("contract_id", contractID),
		ub.Equal("seq", seq),
	))

	sqlq, args := ub.BuildWithFlavor(sqlbuilder.PostgreSQL)
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) FindDue(nowRFC3339 string, graceDays int) ([]sharia.DueItem, error) {
	now, err := time.Parse(time.RFC3339, nowRFC3339)
	if err != nil {
		return nil, err
	}
	cutoff := now.UTC().AddDate(0, 0, -graceDays).Format(time.RFC3339)

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("contract_id", "seq", "due_date", "amount")
	sb.From(s.table("sharia_schedule"))
	sb.Where(sb.And(
		sb.Equal("status", sharia.StatusPending),
		sb.LessEqualThan("due_date", cutoff),
	))
	sb.OrderBy("due_date").Asc()

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []sharia.DueItem{}
	for rows.Next() {
		var it sharia.DueItem
		if err := rows.Scan(&it.ContractID, &it.Seq, &it.DueDate, &it.Amount); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (s *PGStore) AppendAudit(e sharia.AuditEvent) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("sharia_audit"))
	ib.Cols("contract_id", "seq", "event", "transition", "decision", "reason", "standard_ref", "tx_id", "payload", "prev_hash", "hash", "created_at")
	ib.Values(e.ContractID, e.Seq, e.Event, e.Transition, e.Decision, e.Reason, e.StandardRef, e.TxID, e.Payload, e.PrevHash, e.Hash, e.CreatedAt)

	sqlq, args := ib.BuildWithFlavor(sqlbuilder.PostgreSQL)
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) GetAudit(contractID string) ([]sharia.AuditEvent, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("contract_id", "seq", "event", "transition", "decision", "reason", "standard_ref", "tx_id", "payload", "prev_hash", "hash", "created_at")
	sb.From(s.table("sharia_audit"))
	sb.Where(sb.Equal("contract_id", contractID))
	sb.OrderBy("seq").Asc()

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []sharia.AuditEvent{}
	for rows.Next() {
		var e sharia.AuditEvent
		if err := rows.Scan(&e.ContractID, &e.Seq, &e.Event, &e.Transition, &e.Decision, &e.Reason, &e.StandardRef, &e.TxID, &e.Payload, &e.PrevHash, &e.Hash, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (s *PGStore) LastAuditHash(contractID string) (int, string, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("seq", "hash")
	sb.From(s.table("sharia_audit"))
	sb.Where(sb.Equal("contract_id", contractID))
	sb.OrderBy("seq").Desc()
	sb.Limit(1)

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	var seq int
	var hash string
	err := s.Conn().QueryRow(context.Background(), sqlq, args...).Scan(&seq, &hash)
	if err == pgx.ErrNoRows {
		return -1, "", nil
	}
	if err != nil {
		return -1, "", err
	}
	return seq, hash, nil
}
