package postgres

import (
	"context"
	"fmt"

	"github.com/amezianechayer/corren/flows"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v4"
)

var flowCols = []string{"id", "name", "trigger_type", "schedule", "steps", "last_run_at", "created_at", "updated_at"}
var instanceCols = []string{"id", "flow_id", "status", "current_step", "input", "resume_at", "error", "created_at", "updated_at"}

func (s *PGStore) SaveFlow(f flows.Flow) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("flows"))
	ib.Cols(flowCols...)
	ib.Values(f.ID, f.Name, f.Trigger, f.Schedule, string(f.Steps), f.LastRunAt, f.CreatedAt, f.UpdatedAt)
	return s.execWallet(ib)
}

func (s *PGStore) UpdateFlow(f flows.Flow) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("flows"))
	ub.Set(ub.Assign("last_run_at", f.LastRunAt), ub.Assign("updated_at", f.UpdatedAt))
	ub.Where(ub.Equal("id", f.ID))
	return s.execWallet(ub)
}

func pgScanFlow(row interface{ Scan(...interface{}) error }) (flows.Flow, error) {
	var f flows.Flow
	var steps string
	err := row.Scan(&f.ID, &f.Name, &f.Trigger, &f.Schedule, &steps, &f.LastRunAt, &f.CreatedAt, &f.UpdatedAt)
	f.Steps = []byte(steps)
	return f, err
}

func (s *PGStore) GetFlow(id string) (flows.Flow, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(flowCols...)
	sb.From(s.table("flows"))
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	f, err := pgScanFlow(s.Conn().QueryRow(context.Background(), sqlq, args...))
	if err == pgx.ErrNoRows {
		return flows.Flow{}, fmt.Errorf("flow not found: %s", id)
	}
	return f, err
}

func (s *PGStore) ListFlows() ([]flows.Flow, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(flowCols...)
	sb.From(s.table("flows"))
	sb.OrderBy("created_at").Desc()
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []flows.Flow{}
	for rows.Next() {
		f, err := pgScanFlow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

func (s *PGStore) SaveInstance(i flows.FlowInstance) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(s.table("flow_instances"))
	ib.Cols(instanceCols...)
	ib.Values(i.ID, i.FlowID, i.Status, i.CurrentStep, string(i.Input), i.ResumeAt, i.Error, i.CreatedAt, i.UpdatedAt)
	return s.execWallet(ib)
}

func (s *PGStore) UpdateInstance(i flows.FlowInstance) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(s.table("flow_instances"))
	ub.Set(
		ub.Assign("status", i.Status),
		ub.Assign("current_step", i.CurrentStep),
		ub.Assign("resume_at", i.ResumeAt),
		ub.Assign("error", i.Error),
		ub.Assign("updated_at", i.UpdatedAt),
	)
	ub.Where(ub.Equal("id", i.ID))
	return s.execWallet(ub)
}

func pgScanInstance(row interface{ Scan(...interface{}) error }) (flows.FlowInstance, error) {
	var i flows.FlowInstance
	var input string
	err := row.Scan(&i.ID, &i.FlowID, &i.Status, &i.CurrentStep, &input, &i.ResumeAt, &i.Error, &i.CreatedAt, &i.UpdatedAt)
	i.Input = []byte(input)
	return i, err
}

func (s *PGStore) GetInstance(id string) (flows.FlowInstance, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(instanceCols...)
	sb.From(s.table("flow_instances"))
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	i, err := pgScanInstance(s.Conn().QueryRow(context.Background(), sqlq, args...))
	if err == pgx.ErrNoRows {
		return flows.FlowInstance{}, fmt.Errorf("instance not found: %s", id)
	}
	return i, err
}

func (s *PGStore) listInstancesWhere(col, val string) ([]flows.FlowInstance, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(instanceCols...)
	sb.From(s.table("flow_instances"))
	sb.Where(sb.Equal(col, val))
	sb.OrderBy("created_at").Desc()
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.PostgreSQL)
	rows, err := s.Conn().Query(context.Background(), sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []flows.FlowInstance{}
	for rows.Next() {
		i, err := pgScanInstance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, nil
}

func (s *PGStore) ListInstances(flowID string) ([]flows.FlowInstance, error) {
	return s.listInstancesWhere("flow_id", flowID)
}

func (s *PGStore) ListInstancesByStatus(status string) ([]flows.FlowInstance, error) {
	return s.listInstancesWhere("status", status)
}
