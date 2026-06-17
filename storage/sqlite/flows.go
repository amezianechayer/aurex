package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/amezianechayer/corren/flows"
	"github.com/huandu/go-sqlbuilder"
)

var flowCols = []string{"id", "name", "trigger_type", "schedule", "steps", "last_run_at", "created_at", "updated_at"}
var instanceCols = []string{"id", "flow_id", "status", "current_step", "input", "resume_at", "error", "created_at", "updated_at"}

func (s *SQLiteStore) SaveFlow(f flows.Flow) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("flows")
	ib.Cols(flowCols...)
	ib.Values(f.ID, f.Name, f.Trigger, f.Schedule, string(f.Steps), f.LastRunAt, f.CreatedAt, f.UpdatedAt)
	return s.exec(ib)
}

func (s *SQLiteStore) UpdateFlow(f flows.Flow) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("flows")
	ub.Set(ub.Assign("last_run_at", f.LastRunAt), ub.Assign("updated_at", f.UpdatedAt))
	ub.Where(ub.Equal("id", f.ID))
	return s.exec(ub)
}

func scanFlow(row interface{ Scan(...interface{}) error }) (flows.Flow, error) {
	var f flows.Flow
	var steps string
	err := row.Scan(&f.ID, &f.Name, &f.Trigger, &f.Schedule, &steps, &f.LastRunAt, &f.CreatedAt, &f.UpdatedAt)
	f.Steps = []byte(steps)
	return f, err
}

func (s *SQLiteStore) GetFlow(id string) (flows.Flow, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(flowCols...)
	sb.From("flows")
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	f, err := scanFlow(s.db.QueryRow(sqlq, args...))
	if err == sql.ErrNoRows {
		return flows.Flow{}, fmt.Errorf("flow not found: %s", id)
	}
	return f, err
}

func (s *SQLiteStore) ListFlows() ([]flows.Flow, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(flowCols...)
	sb.From("flows")
	sb.OrderBy("created_at").Desc()
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	rows, err := s.db.Query(sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []flows.Flow{}
	for rows.Next() {
		f, err := scanFlow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

func (s *SQLiteStore) SaveInstance(i flows.FlowInstance) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("flow_instances")
	ib.Cols(instanceCols...)
	ib.Values(i.ID, i.FlowID, i.Status, i.CurrentStep, string(i.Input), i.ResumeAt, i.Error, i.CreatedAt, i.UpdatedAt)
	return s.exec(ib)
}

func (s *SQLiteStore) UpdateInstance(i flows.FlowInstance) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("flow_instances")
	ub.Set(
		ub.Assign("status", i.Status),
		ub.Assign("current_step", i.CurrentStep),
		ub.Assign("resume_at", i.ResumeAt),
		ub.Assign("error", i.Error),
		ub.Assign("updated_at", i.UpdatedAt),
	)
	ub.Where(ub.Equal("id", i.ID))
	return s.exec(ub)
}

func scanInstance(row interface{ Scan(...interface{}) error }) (flows.FlowInstance, error) {
	var i flows.FlowInstance
	var input string
	err := row.Scan(&i.ID, &i.FlowID, &i.Status, &i.CurrentStep, &input, &i.ResumeAt, &i.Error, &i.CreatedAt, &i.UpdatedAt)
	i.Input = []byte(input)
	return i, err
}

func (s *SQLiteStore) GetInstance(id string) (flows.FlowInstance, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(instanceCols...)
	sb.From("flow_instances")
	sb.Where(sb.Equal("id", id))
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	i, err := scanInstance(s.db.QueryRow(sqlq, args...))
	if err == sql.ErrNoRows {
		return flows.FlowInstance{}, fmt.Errorf("instance not found: %s", id)
	}
	return i, err
}

func (s *SQLiteStore) listInstancesWhere(col, val string) ([]flows.FlowInstance, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(instanceCols...)
	sb.From("flow_instances")
	sb.Where(sb.Equal(col, val))
	sb.OrderBy("created_at").Desc()
	sqlq, args := sb.BuildWithFlavor(sqlbuilder.SQLite)
	rows, err := s.db.Query(sqlq, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []flows.FlowInstance{}
	for rows.Next() {
		i, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, nil
}

func (s *SQLiteStore) ListInstances(flowID string) ([]flows.FlowInstance, error) {
	return s.listInstancesWhere("flow_id", flowID)
}

func (s *SQLiteStore) ListInstancesByStatus(status string) ([]flows.FlowInstance, error) {
	return s.listInstancesWhere("status", status)
}
