package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amezianechayer/corren/guard"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/viper"
)

func (s *PGStore) SaveRule(r guard.Rule) error {
	params := string(r.Params)
	if params == "" {
		params = "{}"
	}

	enabledInt := 0
	if r.Enabled {
		enabledInt = 1
	}

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("guard_rules"))
	ib.Cols("id", "kind", "params", "action", "reason", "standard_ref", "enabled", "created_at", "updated_at")
	ib.Values(r.ID, r.Kind, params, r.Action, r.Reason, r.StandardRef, enabledInt, r.CreatedAt, r.UpdatedAt)

	sqlq, args := ib.BuildWithFlavor(sqlbuilder.PostgreSQL)
	if viper.GetBool("debug") {
		fmt.Println(sqlq, args)
	}

	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) UpdateRule(r guard.Rule) error {
	params := string(r.Params)
	if params == "" {
		params = "{}"
	}

	enabledInt := 0
	if r.Enabled {
		enabledInt = 1
	}

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("guard_rules"))
	ub.Set(
		ub.Assign("kind", r.Kind),
		ub.Assign("params", params),
		ub.Assign("action", r.Action),
		ub.Assign("reason", r.Reason),
		ub.Assign("standard_ref", r.StandardRef),
		ub.Assign("enabled", enabledInt),
		ub.Assign("updated_at", r.UpdatedAt),
	)
	ub.Where(ub.Equal("id", r.ID))

	sqlq, args := ub.BuildWithFlavor(sqlbuilder.PostgreSQL)
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) GetRule(id string) (guard.Rule, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "kind", "params", "action", "reason", "standard_ref", "enabled", "created_at", "updated_at")
	sb.From(s.table("guard_rules"))
	sb.Where(sb.Equal("id", id))

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)

	var r guard.Rule
	var params string
	var enabledInt int

	err := s.Conn().QueryRow(context.Background(), sqlq, args...).Scan(
		&r.ID, &r.Kind, &params, &r.Action, &r.Reason, &r.StandardRef, &enabledInt, &r.CreatedAt, &r.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return guard.Rule{}, &guard.Error{Code: guard.ErrNotFound, Message: "rule not found"}
	}
	if err != nil {
		return guard.Rule{}, err
	}

	r.Params = json.RawMessage(params)
	r.Enabled = enabledInt == 1
	return r, nil
}

func (s *PGStore) ListRules() ([]guard.Rule, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "kind", "params", "action", "reason", "standard_ref", "enabled", "created_at", "updated_at")
	sb.From(s.table("guard_rules"))
	sb.OrderBy("created_at").Asc()

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := []guard.Rule{}
	for rows.Next() {
		var r guard.Rule
		var params string
		var enabledInt int
		if err := rows.Scan(&r.ID, &r.Kind, &params, &r.Action, &r.Reason, &r.StandardRef, &enabledInt, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Params = json.RawMessage(params)
		r.Enabled = enabledInt == 1
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *PGStore) DeleteRule(id string) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(s.table("guard_rules"))
	db.Where(db.Equal("id", id))

	sqlq, args := db.BuildWithFlavor(sqlbuilder.PostgreSQL)
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) AppendGuardEvent(e guard.GuardEvent) error {
	var count int64
	if err := s.Conn().QueryRow(context.Background(),
		fmt.Sprintf(`SELECT COUNT(*) FROM %s`, s.table("guard_events")),
	).Scan(&count); err != nil {
		return err
	}
	e.ID = count + 1

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("guard_events"))
	ib.Cols("id", "rule_id", "action", "reason", "standard_ref", "tx_reference", "payload", "created_at")
	ib.Values(e.ID, e.RuleID, e.Action, e.Reason, e.StandardRef, e.TxReference, e.Payload, e.CreatedAt)

	sqlq, args := ib.BuildWithFlavor(sqlbuilder.PostgreSQL)
	_, err := s.Conn().Exec(context.Background(), sqlq, args...)
	return err
}

func (s *PGStore) ListGuardEvents(limit, offset int) ([]guard.GuardEvent, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "rule_id", "action", "reason", "standard_ref", "tx_reference", "payload", "created_at")
	sb.From(s.table("guard_events"))
	sb.OrderBy("id").Asc()
	sb.Limit(limit)
	sb.Offset(offset)

	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []guard.GuardEvent{}
	for rows.Next() {
		var ev guard.GuardEvent
		if err := rows.Scan(&ev.ID, &ev.RuleID, &ev.Action, &ev.Reason, &ev.StandardRef, &ev.TxReference, &ev.Payload, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, nil
}
