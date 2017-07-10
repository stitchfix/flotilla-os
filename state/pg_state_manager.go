package state

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"strings"
)

type SQLStateManager struct {
	db *sqlx.DB
}

func (sm *SQLStateManager) Initialize(dburl string) error {
	var err error
	if sm.db, err = sqlx.Connect("postgres", dburl); err != nil {
		return err
	}
	if err = sm.createTables(); err != nil {
		return err
	}
	return nil
}

func (sm *SQLStateManager) createTables() error {
	_, err := sm.db.Exec(CreateTablesSQL)
	return err
}

func (sm *SQLStateManager) makeWhereClause(filters map[string]string) []string {
	wc := make([]string, len(filters))
	i := 0
	for k, v := range filters {
		fmtString := "%s='%s'"
		if k == "image" || k == "alias" {
			fmtString = "%s like '%%%s%%'"
		}
		wc[i] = fmt.Sprintf(fmtString, k, v)
		i++
	}
	return wc
}

func (sm *SQLStateManager) makeEnvWhereClause(filters map[string]string) []string {
	wc := make([]string, len(filters))
	i := 0
	for k, v := range filters {
		fmtString := `env::jsonb @> '[{"name":"%s","value":"%s"}]'`
		wc[i] = fmt.Sprintf(fmtString, k, v)
		i++
	}

	return wc
}

func (sm *SQLStateManager) orderBy(obj Orderable, field string, order string) (string, error) {
	if order == "asc" || order == "desc" {
		if obj.ValidOrderField(field) {
			return fmt.Sprintf("order by %s %s", field, order), nil
		} else {
			return "", errors.New(fmt.Sprintf("Invalid field to order by [%s], must be one of [%s]",
				field,
				strings.Join(obj.ValidOrderFields(), ", ")))
		}
	} else {
		return "", errors.New(fmt.Sprintf("Invalid order string, must be one of ('asc', 'desc'), was %s", order))
	}
}

//
// Definitions
//

func (sm *SQLStateManager) ListDefinitions(
	limit int, offset int, sortBy string,
	order string, filters map[string]string,
	envFilters map[string]string) (DefinitionList, error) {

	var err error
	var result DefinitionList
	var whereClause, orderQuery string
	where := append(sm.makeWhereClause(filters), sm.makeEnvWhereClause(envFilters)...)
	if len(where) > 0 {
		whereClause = fmt.Sprintf("where %s", strings.Join(where, " and "))
	}

	orderQuery, err = sm.orderBy(&Definition{}, sortBy, order)
	if err != nil {
		return result, err
	}

	sql := fmt.Sprintf(ListDefinitionsSQL, whereClause, orderQuery)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Definitions, sql, limit, offset)
	if err != nil {
		return result, err
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (sm *SQLStateManager) GetDefinition(definitionID string) (Definition, error) {
	var err error
	var definition Definition
	err = sm.db.Get(&definition, GetDefinitionSQL, definitionID)
	return definition, err
}

func (sm *SQLStateManager) UpdateDefinition(definitionID string, updates Definition) error {
	var err error
	existing, err := sm.GetDefinition(definitionID)
	if err != nil {
		return err
	}

	existing.updateWith(updates)

	selectForUpdate := `SELECT * FROM task_def WHERE definition_id = $1 FOR UPDATE;`
	deleteEnv := `DELETE FROM task_def_environments WHERE task_def_id = $1;`
	deletePorts := `DELETE FROM task_def_ports WHERE task_def_id = $1;`
	update := `
    UPDATE task_def SET
      arn = $2, image = $3,
      container_name = $4, "user" = $5,
      alias = $6, memory = $7,
      command = $8
    WHERE definition_id = $1;
    `

	insertEnv := `
    INSERT INTO task_def_environments(
      task_def_id, name, value
    )
    VALUES ($1, $2, $3);
    `

	insertPorts := `
    INSERT INTO task_def_ports(
      task_def_id, port
    ) VALUES ($1, $2);
    `

	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(selectForUpdate, definitionID); err != nil {
		return err
	}

	if _, err = tx.Exec(deleteEnv, definitionID); err != nil {
		return err
	}

	if _, err = tx.Exec(deletePorts, definitionID); err != nil {
		return err
	}

	if _, err = tx.Exec(
		update, definitionID,
		existing.Arn, existing.Image, existing.ContainerName,
		existing.User, existing.Alias, existing.Memory,
		existing.Command); err != nil {
		return err
	}

	if existing.Env != nil {
		for _, e := range *existing.Env {
			if _, err = tx.Exec(insertEnv, definitionID, e.Name, e.Value); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if existing.Ports != nil {
		for _, p := range *existing.Ports {
			if _, err = tx.Exec(insertPorts, definitionID, p); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (sm *SQLStateManager) CreateDefinition(d Definition) error {
	var err error
	insert := `
    INSERT INTO task_def(
      arn, definition_id, image, group_name,
      container_name, "user", alias, memory, command
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
    `

	insertEnv := `
    INSERT INTO task_def_environments(
      task_def_id, name, value
    )
    VALUES ($1, $2, $3);
    `

	insertPorts := `
    INSERT INTO task_def_ports(
      task_def_id, port
    ) VALUES ($1, $2);
    `
	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(insert,
		d.Arn, d.DefinitionID, d.Image, d.GroupName, d.ContainerName,
		d.User, d.Alias, d.Memory, d.Command); err != nil {
		tx.Rollback()
		return err
	}

	if d.Env != nil {
		for _, e := range *d.Env {
			if _, err = tx.Exec(insertEnv, d.DefinitionID, e.Name, e.Value); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if d.Ports != nil {
		for _, p := range *d.Ports {
			if _, err = tx.Exec(insertPorts, d.DefinitionID, p); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

func (sm *SQLStateManager) DeleteDefinition(definitionID string) error {
	var err error

	delTaskEnvs := `
    DELETE FROM task_environments WHERE task_id in (
      SELECT run_id as task_id from task WHERE definition_id = $1
    )
    `

	statements := []string{
		"DELETE FROM task_def_environments WHERE task_def_id = $1",
		"DELETE FROM task_def_ports WHERE task_def_id = $1",
		delTaskEnvs,
		"DELETE FROM task WHERE definition_id = $1",
		"DELETE FROM task_def WHERE definition_id = $1",
	}
	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	for _, stmt := range statements {
		if _, err = tx.Exec(stmt, definitionID); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

//
// Runs
//

func (sm *SQLStateManager) ListRuns(
	limit int, offset int, sortBy string,
	order string, filters map[string]string,
	envFilters map[string]string) (RunList, error) {

	var err error
	var result RunList
	var whereClause, orderQuery string
	where := append(sm.makeWhereClause(filters), sm.makeEnvWhereClause(envFilters)...)
	if len(where) > 0 {
		whereClause = fmt.Sprintf("where %s", strings.Join(where, " and "))
	}

	orderQuery, err = sm.orderBy(&Run{}, sortBy, order)
	if err != nil {
		return result, err
	}

	sql := fmt.Sprintf(ListRunsSQL, whereClause, orderQuery)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Runs, sql, limit, offset)
	if err != nil {
		return result, err
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (sm *SQLStateManager) GetRun(runID string) (Run, error) {
	var err error
	var r Run
	err = sm.db.Get(&r, GetRunSQL, runID)
	return r, err
}

func (sm *SQLStateManager) UpdateRun(runID string, updates Run) error {
	var err error
	existing, err := sm.GetRun(runID)
	if err != nil {
		return err
	}

	existing.updateWith(updates)

	selectForUpdate := `SELECT * FROM task WHERE run_id = $1 FOR UPDATE;`
	deleteEnv := `DELETE FROM task_environments WHERE task_id = $1;`
	update := `
    UPDATE task SET
      task_arn = $2, definition_id = $3,
      cluster_name = $4, exit_code = $5,
      status = $6, started_at = $7,
      finished_at = $8, instance_id = $9,
      instance_dns_name = $10,
      group_name = $11
    WHERE run_id = $1;
    `

	insertEnv := `
    INSERT INTO task_environments(
      task_id, name, value
    )
    VALUES ($1, $2, $3);
    `

	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(selectForUpdate, runID); err != nil {
		return err
	}

	if _, err = tx.Exec(deleteEnv, runID); err != nil {
		return err
	}

	if _, err = tx.Exec(
		update, runID,
		existing.TaskArn, existing.DefinitionID,
		existing.ClusterName, existing.ExitCode,
		existing.Status, existing.StartedAt,
		existing.FinishedAt, existing.InstanceID,
		existing.InstanceDNSName, existing.GroupName); err != nil {
		return err
	}

	if existing.Env != nil {
		for _, e := range *existing.Env {
			if _, err = tx.Exec(insertEnv, runID, e.Name, e.Value); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

func (sm *SQLStateManager) CreateRun(r Run) error {
	var err error
	insert := `
	INSERT INTO task (
      task_arn, run_id, definition_id, cluster_name, exit_code, status,
      started_at, finished_at, instance_id, instance_dns_name, group_name
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
    );
    `

	insertEnv := `
    INSERT INTO task_environments(
      task_id, name, value
    )
    VALUES ($1, $2, $3);
    `

	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(insert,
		r.TaskArn, r.RunID, r.DefinitionID,
		r.ClusterName, r.ExitCode, r.Status,
		r.StartedAt, r.FinishedAt,
		r.InstanceID, r.InstanceDNSName, r.GroupName); err != nil {
		tx.Rollback()
		return err
	}

	if r.Env != nil {
		for _, e := range *r.Env {
			if _, err = tx.Exec(insertEnv, r.RunID, e.Name, e.Value); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

func (sm *SQLStateManager) Cleanup() error {
	return sm.db.Close()
}

type Orderable interface {
	ValidOrderField(field string) bool
	ValidOrderFields() []string
}

func (d *Definition) ValidOrderField(field string) bool {
	for _, f := range d.ValidOrderFields() {
		if field == f {
			return true
		}
	}
	return false
}

func (d *Definition) ValidOrderFields() []string {
	return []string{"alias", "image", "group_name", "memory"}
}

func (r *Run) ValidOrderField(field string) bool {
	for _, f := range r.ValidOrderFields() {
		if field == f {
			return true
		}
	}
	return false
}

func (r *Run) ValidOrderFields() []string {
	return []string{"run_id", "cluster_name", "status", "started_at", "finished_at", "group_name"}
}

func (e *EnvList) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

func (e EnvList) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

func (e *PortsList) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}
func (e PortsList) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}
