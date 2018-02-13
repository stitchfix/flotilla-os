package state

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	// Pull in postgres specific drivers
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"math"
	"strings"
	"time"
)

//
// SQLStateManager uses postgresql to manage state
//
type SQLStateManager struct {
	db *sqlx.DB
}

//
// Name is the name of the state manager - matches value in configuration
//
func (sm *SQLStateManager) Name() string {
	return "postgres"
}

//
// Initialize creates tables if they do not exist
//
func (sm *SQLStateManager) Initialize(conf config.Config) error {
	dburl := conf.GetString("database_url")
	createSchema := conf.GetBool("create_database_schema")

	var err error
	if sm.db, err = sqlx.Open("postgres", dburl); err != nil {
		return err
	}

	if createSchema {
		// Since this happens at initialization we
		// could encounter racy conditions waiting for pg
		// to become available. Wait for it a bit
		if err = sm.db.Ping(); err != nil {
			// Try 3 more times
			// 5, 10, 20
			for i := 0; i < 3 && err != nil; i++ {
				time.Sleep(time.Duration(5*math.Pow(2, float64(i))) * time.Second)
				err = sm.db.Ping()
			}
			if err != nil {
				return fmt.Errorf("error trying to connect to postgres, retries exhausted; error: [%s]", err.Error())
			}
		}

		if err = sm.createTables(); err != nil {
			return err
		}
	}
	return nil
}

func (sm *SQLStateManager) createTables() error {
	_, err := sm.db.Exec(CreateTablesSQL)
	return err
}

func (sm *SQLStateManager) makeWhereClause(filters map[string][]string) []string {

	// These will be joined with "AND"
	wc := []string{}
	for k, v := range filters {
		if len(v) > 1 {
			// No like queries for multiple filters with same key
			quoted := make([]string, len(v))
			for i, filterVal := range v {
				quoted[i] = fmt.Sprintf("'%s'", filterVal)
			}
			wc = append(wc, fmt.Sprintf("%s in (%s)", k, strings.Join(quoted, ",")))
		} else if len(v) == 1 {
			fmtString := "%s='%s'"
			fieldName := k
			if k == "image" || k == "alias" || k == "group_name" || k == "command" || k == "text" {
				fmtString = "%s like '%%%s%%'"
			} else if strings.HasSuffix(k, "_since") {
				fieldName = strings.Replace(k, "_since", "", -1)
				fmtString = "%s > '%s'"
			} else if strings.HasSuffix(k, "_until") {
				fieldName = strings.Replace(k, "_until", "", -1)
				fmtString = "%s < '%s'"
			}
			wc = append(wc, fmt.Sprintf(fmtString, fieldName, v[0]))
		}
	}
	return wc
}

func (sm *SQLStateManager) makeEnvWhereClause(filters map[string]string) []string {
	wc := make([]string, len(filters))
	i := 0
	for k, v := range filters {
		fmtString := `env @> '[{"name":"%s","value":"%s"}]'`
		wc[i] = fmt.Sprintf(fmtString, k, v)
		i++
	}

	return wc
}

func (sm *SQLStateManager) orderBy(obj orderable, field string, order string) (string, error) {
	if order == "asc" || order == "desc" {
		if obj.validOrderField(field) {
			return fmt.Sprintf("order by %s %s NULLS LAST", field, order), nil
		}
		return "", fmt.Errorf("Invalid field to order by [%s], must be one of [%s]",
			field,
			strings.Join(obj.validOrderFields(), ", "))
	}
	return "", fmt.Errorf("Invalid order string, must be one of ('asc', 'desc'), was %s", order)
}

//
// ListDefinitions returns a DefinitionList
// limit: limit the result to this many definitions
// offset: start the results at this offset
// sortBy: sort by this field
// order: 'asc' or 'desc'
// filters: map of field filters on Definition - joined with AND
// envFilters: map of environment variable filters - joined with AND
//
func (sm *SQLStateManager) ListDefinitions(
	limit int, offset int, sortBy string,
	order string, filters map[string][]string,
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

//
// GetDefinition returns a single definition by id
//
func (sm *SQLStateManager) GetDefinition(definitionID string) (Definition, error) {
	var err error
	var definition Definition
	err = sm.db.Get(&definition, GetDefinitionSQL, definitionID)
	if err != nil && err == sql.ErrNoRows {
		return definition, exceptions.MissingResource{
			fmt.Sprintf("Definition with ID %s not found", definitionID)}
	}
	return definition, err
}

//
// GetDefinitionByAlias returns a single definition by id
//
func (sm *SQLStateManager) GetDefinitionByAlias(alias string) (Definition, error) {
	var err error
	var definition Definition
	err = sm.db.Get(&definition, GetDefinitionByAliasSQL, alias)
	if err != nil && err == sql.ErrNoRows {
		return definition, exceptions.MissingResource{
			fmt.Sprintf("Definition with alias %s not found", alias)}
	}
	return definition, err
}

//
// UpdateDefinition updates a definition
// - updates can be partial
//
func (sm *SQLStateManager) UpdateDefinition(definitionID string, updates Definition) (Definition, error) {
	var (
		err      error
		existing Definition
	)
	existing, err = sm.GetDefinition(definitionID)
	if err != nil {
		return existing, err
	}

	existing.UpdateWith(updates)

	selectForUpdate := `SELECT * FROM task_def WHERE definition_id = $1 FOR UPDATE;`
	deletePorts := `DELETE FROM task_def_ports WHERE task_def_id = $1;`
	deleteTags := `DELETE FROM task_def_tags WHERE task_def_id = $1`
	update := `
    UPDATE task_def SET
      arn = $2, image = $3,
      container_name = $4, "user" = $5,
      alias = $6, memory = $7,
      command = $8, env = $9
    WHERE definition_id = $1;
    `

	insertPorts := `
    INSERT INTO task_def_ports(
      task_def_id, port
    ) VALUES ($1, $2);
    `

	insertDefTags := `
	INSERT INTO task_def_tags(
	  task_def_id, tag_id
	) VALUES ($1, $2);
	`

	insertTags := `
	INSERT INTO tags(text) SELECT $1 WHERE NOT EXISTS (SELECT text from tags where text = $2)
	`

	tx, err := sm.db.Begin()
	if err != nil {
		return existing, err
	}

	if _, err = tx.Exec(selectForUpdate, definitionID); err != nil {
		return existing, err
	}

	if _, err = tx.Exec(deletePorts, definitionID); err != nil {
		return existing, err
	}

	if _, err = tx.Exec(deleteTags, definitionID); err != nil {
		return existing, err
	}

	if _, err = tx.Exec(
		update, definitionID,
		existing.Arn, existing.Image, existing.ContainerName,
		existing.User, existing.Alias, existing.Memory,
		existing.Command, existing.Env); err != nil {
		return existing, err
	}

	if existing.Ports != nil {
		for _, p := range *existing.Ports {
			if _, err = tx.Exec(insertPorts, definitionID, p); err != nil {
				tx.Rollback()
				return existing, err
			}
		}
	}

	if existing.Tags != nil {
		for _, t := range *existing.Tags {
			if _, err = tx.Exec(insertTags, t, t); err != nil {
				tx.Rollback()
				return existing, err
			}
			if _, err = tx.Exec(insertDefTags, definitionID, t); err != nil {
				tx.Rollback()
				return existing, err
			}
		}
	}
	return existing, tx.Commit()
}

//
// CreateDefinition creates the passed in definition object
// - error if definition already exists
//
func (sm *SQLStateManager) CreateDefinition(d Definition) error {
	var err error
	insert := `
    INSERT INTO task_def(
      arn, definition_id, image, group_name,
      container_name, "user", alias, memory, command, env
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
    `

	insertPorts := `
    INSERT INTO task_def_ports(
      task_def_id, port
    ) VALUES ($1, $2);
    `

	insertDefTags := `
	INSERT INTO task_def_tags(
	  task_def_id, tag_id
	) VALUES ($1, $2);
	`

	insertTags := `
	INSERT INTO tags(text) SELECT $1 WHERE NOT EXISTS (SELECT text from tags where text = $2)
	`

	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(insert,
		d.Arn, d.DefinitionID, d.Image, d.GroupName, d.ContainerName,
		d.User, d.Alias, d.Memory, d.Command, d.Env); err != nil {
		tx.Rollback()
		return err
	}

	if d.Ports != nil {
		for _, p := range *d.Ports {
			if _, err = tx.Exec(insertPorts, d.DefinitionID, p); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if d.Tags != nil {
		for _, t := range *d.Tags {
			if _, err = tx.Exec(insertTags, t, t); err != nil {
				tx.Rollback()
				return err
			}
			if _, err = tx.Exec(insertDefTags, d.DefinitionID, t); err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit()
}

//
// DeleteDefinition deletes definition and associated runs and environment variables
//
func (sm *SQLStateManager) DeleteDefinition(definitionID string) error {
	var err error

	statements := []string{
		"DELETE FROM task_def_ports WHERE task_def_id = $1",
		"DELETE FROM task_def_tags WHERE task_def_id = $1",
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
// ListRuns returns a RunList
// limit: limit the result to this many runs
// offset: start the results at this offset
// sortBy: sort by this field
// order: 'asc' or 'desc'
// filters: map of field filters on Run - joined with AND
// envFilters: map of environment variable filters - joined with AND
//
func (sm *SQLStateManager) ListRuns(
	limit int, offset int, sortBy string,
	order string, filters map[string][]string,
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

//
// GetRun gets run by id
//
func (sm *SQLStateManager) GetRun(runID string) (Run, error) {
	var err error
	var r Run
	err = sm.db.Get(&r, GetRunSQL, runID)
	return r, err
}

//
// UpdateRun updates run with updates - can be partial
//
func (sm *SQLStateManager) UpdateRun(runID string, updates Run) (Run, error) {
	var (
		err      error
		existing Run
	)

	tx, err := sm.db.Begin()
	if err != nil {
		return existing, err
	}

	rows, err := tx.Query(GetRunSQLForUpdate, runID)
	if err != nil {
		tx.Rollback()
		return existing, err
	}

	for rows.Next() {
		err = rows.Scan(
			&existing.TaskArn, &existing.RunID, &existing.DefinitionID, &existing.Alias, &existing.Image,
			&existing.ClusterName, &existing.ExitCode, &existing.Status, &existing.StartedAt,
			&existing.FinishedAt, &existing.InstanceID, &existing.InstanceDNSName, &existing.GroupName,
			&existing.User, &existing.TaskType, &existing.Env)
	}
	if err != nil {
		return existing, err
	}

	existing.UpdateWith(updates)

	update := `
    UPDATE task SET
      task_arn = $2, definition_id = $3,
	  alias = $4, image = $5,
      cluster_name = $6, exit_code = $7,
      status = $8, started_at = $9,
      finished_at = $10, instance_id = $11,
      instance_dns_name = $12,
      group_name = $13, env = $14
    WHERE run_id = $1;
    `

	if _, err = tx.Exec(
		update, runID,
		existing.TaskArn, existing.DefinitionID,
		existing.Alias, existing.Image,
		existing.ClusterName, existing.ExitCode,
		existing.Status, existing.StartedAt,
		existing.FinishedAt, existing.InstanceID,
		existing.InstanceDNSName, existing.GroupName,
		existing.Env); err != nil {
		tx.Rollback()
		return existing, err
	}

	return existing, tx.Commit()
}

//
// CreateRun creates the passed in run
//
func (sm *SQLStateManager) CreateRun(r Run) error {
	var err error
	insert := `
	INSERT INTO task (
      task_arn, run_id, definition_id, alias, image, cluster_name, exit_code, status,
      started_at, finished_at, instance_id, instance_dns_name, group_name,
      env, task_type
    ) VALUES (
      $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 'task'
    );
    `

	tx, err := sm.db.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(insert,
		r.TaskArn, r.RunID, r.DefinitionID,
		r.Alias, r.Image, r.ClusterName,
		r.ExitCode, r.Status, r.StartedAt,
		r.FinishedAt, r.InstanceID,
		r.InstanceDNSName, r.GroupName, r.Env); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

//
// Metadata
//
func (sm *SQLStateManager) ListGroups(limit int, offset int, name *string) (GroupsList, error) {
	var (
		err         error
		result      GroupsList
		whereClause string
	)
	if name != nil && len(*name) > 0 {
		whereClause = fmt.Sprintf("where %s", strings.Join(
			sm.makeWhereClause(map[string][]string{"group_name": {*name}}), " and "))
	}

	sql := fmt.Sprintf(ListGroupsSQL, whereClause)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Groups, sql, limit, offset)
	if err != nil {
		return result, err
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (sm *SQLStateManager) ListTags(limit int, offset int, name *string) (TagsList, error) {
	var (
		err         error
		result      TagsList
		whereClause string
	)
	if name != nil && len(*name) > 0 {
		whereClause = fmt.Sprintf("where %s", strings.Join(
			sm.makeWhereClause(map[string][]string{"text": {*name}}), " and "))
	}

	sql := fmt.Sprintf(ListTagsSQL, whereClause)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Tags, sql, limit, offset)
	if err != nil {
		return result, err
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, err
	}

	return result, nil
}

//
// Cleanup close any open resources
//
func (sm *SQLStateManager) Cleanup() error {
	return sm.db.Close()
}

type orderable interface {
	validOrderField(field string) bool
	validOrderFields() []string
}

func (d *Definition) validOrderField(field string) bool {
	for _, f := range d.validOrderFields() {
		if field == f {
			return true
		}
	}
	return false
}

func (d *Definition) validOrderFields() []string {
	return []string{"alias", "image", "group_name", "memory"}
}

func (r *Run) validOrderField(field string) bool {
	for _, f := range r.validOrderFields() {
		if field == f {
			return true
		}
	}
	return false
}

func (r *Run) validOrderFields() []string {
	return []string{"run_id", "cluster_name", "status", "started_at", "finished_at", "group_name"}
}

// Scan from db
func (e *EnvList) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e EnvList) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

// Scan from db
func (e *PortsList) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e PortsList) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

// Scan from db
func (e *Tags) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e Tags) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}
