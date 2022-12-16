package state

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/stitchfix/flotilla-os/clients/metrics"

	"github.com/jmoiron/sqlx"

	// Pull in postgres specific drivers
	"database/sql"
	"math"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stitchfix/flotilla-os/config"
	"github.com/stitchfix/flotilla-os/exceptions"
	"go.uber.org/multierr"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	sqlxtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/jmoiron/sqlx"
)

// SQLStateManager uses postgresql to manage state
type SQLStateManager struct {
	db         *sqlx.DB
	readonlyDB *sqlx.DB
}

func (sm *SQLStateManager) ListFailingNodes() (NodeList, error) {
	var err error
	var nodeList NodeList

	err = sm.readonlyDB.Select(&nodeList, ListFailingNodesSQL)

	if err != nil {
		if err == sql.ErrNoRows {
			return nodeList, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Error fetching node list")}
		} else {
			return nodeList, errors.Wrapf(err, "Error fetching node list")
		}
	}
	return nodeList, err
}

func (sm *SQLStateManager) GetPodReAttemptRate() (float32, error) {
	var err error
	attemptRate := float32(1.0)
	err = sm.readonlyDB.Get(&attemptRate, PodReAttemptRate)

	if err != nil {
		if err == sql.ErrNoRows {
			return attemptRate, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Error fetching attempt rate")}
		} else {
			return attemptRate, errors.Wrapf(err, "Error fetching attempt rate")
		}
	}
	return attemptRate, err
}

func (sm *SQLStateManager) GetNodeLifecycle(executableID string, commandHash string) (string, error) {
	var err error
	nodeType := "spot"
	err = sm.readonlyDB.Get(&nodeType, TaskResourcesExecutorNodeLifecycleSQL, executableID, commandHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return nodeType, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Error fetching node type")}
		} else {
			return nodeType, errors.Wrapf(err, "Error fetching node type")
		}
	}
	return nodeType, err
}

func (sm *SQLStateManager) GetTaskHistoricalRuntime(executableID string, runID string) (float32, error) {
	var err error
	minutes := float32(1.0)
	err = sm.readonlyDB.Get(&minutes, TaskExecutionRuntimeCommandSQL, executableID, runID)

	if err != nil {
		if err == sql.ErrNoRows {
			return minutes, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Error fetching TaskRuntime rate")}
		} else {
			return minutes, errors.Wrapf(err, "Error fetching attempt rate")
		}
	}
	return minutes, err
}

func (sm *SQLStateManager) EstimateRunResources(executableID string, runID string) (TaskResources, error) {
	var err error
	var taskResources TaskResources

	err = sm.readonlyDB.Get(&taskResources, TaskResourcesSelectCommandSQL, executableID, runID)

	if err != nil {
		if err == sql.ErrNoRows {
			return taskResources, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Resource usage with executable %s not found", executableID)}
		} else {
			return taskResources, errors.Wrapf(err, "issue getting resources with executable [%s]", executableID)
		}
	}
	return taskResources, err
}

func (sm *SQLStateManager) EstimateExecutorCount(executableID string, commandHash string) (int64, error) {
	var err error
	executorCount := int64(25)
	err = sm.readonlyDB.Get(&executorCount, TaskResourcesExecutorCountSQL, executableID, commandHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return executorCount, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Resource usage with executable %s not found", executableID)}
		} else {
			return executorCount, errors.Wrapf(err, "issue getting resources with executable [%s]", executableID)
		}
	}
	return executorCount, err
}
func (sm *SQLStateManager) CheckIdempotenceKey(idempotenceKey string) (string, error) {
	var err error
	runId := ""
	err = sm.readonlyDB.Get(&runId, TaskIdempotenceKeyCheckSQL, idempotenceKey)

	if err != nil || len(runId) == 0 {
		err = errors.New("no run_id found for idempotence key")
	}
	return runId, err
}

func (sm *SQLStateManager) ExecutorOOM(executableID string, commandHash string) (bool, error) {
	var err error
	executorOOM := false
	err = sm.readonlyDB.Get(&executorOOM, TaskResourcesExecutorOOMSQL, executableID, commandHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return executorOOM, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Resource oom for executable %s not found", executableID)}
		} else {
			return executorOOM, errors.Wrapf(err, "issue getting resources with executable [%s]", executableID)
		}
	}
	return executorOOM, err
}

func (sm *SQLStateManager) DriverOOM(executableID string, commandHash string) (bool, error) {
	var err error
	driverOOM := false
	err = sm.readonlyDB.Get(&driverOOM, TaskResourcesDriverOOMSQL, executableID, commandHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return driverOOM, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Resource oom for driver %s not found", executableID)}
		} else {
			return driverOOM, errors.Wrapf(err, "issue getting resources with executable [%s]", executableID)
		}
	}
	return driverOOM, err
}

// Name is the name of the state manager - matches value in configuration
func (sm *SQLStateManager) Name() string {
	return "postgres"
}

// likeFields are the set of fields
// that are filtered using a `like` clause
var likeFields = map[string]bool{
	"image":       true,
	"alias":       true,
	"group_name":  true,
	"command":     true,
	"text":        true,
	"exit_reason": true,
}

// Initialize creates tables if they do not exist
func (sm *SQLStateManager) Initialize(conf config.Config) error {
	dburl := conf.GetString("database_url")
	readonlyDbUrl := conf.GetString("readonly_database_url")

	createSchema := conf.GetBool("create_database_schema")
	sqltrace.Register("postgres", &pq.Driver{}, sqltrace.WithServiceName("flotilla"))
	var err error
	if sm.db, err = sqlxtrace.Open("postgres", dburl); err != nil {
		return errors.Wrap(err, "unable to open postgres db")
	}

	sqltrace.Register("postgres", &pq.Driver{}, sqltrace.WithServiceName("flotilla"))
	if sm.readonlyDB, err = sqlxtrace.Open("postgres", readonlyDbUrl); err != nil {
		return errors.Wrap(err, "unable to open readonly postgres db")
	}

	if conf.IsSet("database_max_idle_connections") {
		sm.db.SetMaxIdleConns(conf.GetInt("database_max_idle_connections"))
		sm.readonlyDB.SetMaxIdleConns(conf.GetInt("database_max_idle_connections"))
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
				return errors.Wrap(err, "error trying to connect to postgres db, retries exhausted")
			}
		}

		// Populate worker table
		if err = sm.initWorkerTable(conf); err != nil {
			return errors.Wrap(err, "problem populating worker table sql")
		}
	}
	return nil
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
			if likeFields[k] {
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

func (sm *SQLStateManager) orderBy(obj IOrderable, field string, order string) (string, error) {
	if order == "asc" || order == "desc" {
		if obj.ValidOrderField(field) {
			return fmt.Sprintf("order by %s %s NULLS LAST", field, order), nil
		}
		return "", errors.Errorf("Invalid field to order by [%s], must be one of [%s]",
			field,
			strings.Join(obj.ValidOrderFields(), ", "))
	}
	return "", errors.Errorf("Invalid order string, must be one of ('asc', 'desc'), was %s", order)
}

// ListDefinitions returns a DefinitionList
// limit: limit the result to this many definitions
// offset: start the results at this offset
// sortBy: sort by this field
// order: 'asc' or 'desc'
// filters: map of field filters on Definition - joined with AND
// envFilters: map of environment variable filters - joined with AND
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
		return result, errors.WithStack(err)
	}

	sql := fmt.Sprintf(ListDefinitionsSQL, whereClause, orderQuery)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Definitions, sql, limit, offset)
	if err != nil {
		return result, errors.Wrap(err, "issue running list definitions sql")
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list definitions count sql")
	}

	return result, nil
}

// GetDefinition returns a single definition by id
func (sm *SQLStateManager) GetDefinition(definitionID string) (Definition, error) {
	var err error
	var definition Definition
	err = sm.db.Get(&definition, GetDefinitionSQL, definitionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return definition, exceptions.MissingResource{
				fmt.Sprintf("Definition with ID %s not found", definitionID)}
		} else {
			return definition, errors.Wrapf(err, "issue getting definition with id [%s]", definitionID)
		}
	}
	return definition, nil
}

// GetDefinitionByAlias returns a single definition by id
func (sm *SQLStateManager) GetDefinitionByAlias(alias string) (Definition, error) {
	var err error
	var definition Definition
	err = sm.db.Get(&definition, GetDefinitionByAliasSQL, alias)
	if err != nil {
		if err == sql.ErrNoRows {
			return definition, exceptions.MissingResource{
				fmt.Sprintf("Definition with alias %s not found", alias)}
		} else {
			return definition, errors.Wrapf(err, "issue getting definition with alias [%s]", alias)
		}
	}
	return definition, err
}

// UpdateDefinition updates a definition
// - updates can be partial
func (sm *SQLStateManager) UpdateDefinition(definitionID string, updates Definition) (Definition, error) {
	var (
		err      error
		existing Definition
	)
	existing, err = sm.GetDefinition(definitionID)
	if err != nil {
		return existing, errors.WithStack(err)
	}

	existing.UpdateWith(updates)

	selectForUpdate := `SELECT * FROM task_def WHERE definition_id = $1 FOR UPDATE;`
	deletePorts := `DELETE FROM task_def_ports WHERE task_def_id = $1;`
	deleteTags := `DELETE FROM task_def_tags WHERE task_def_id = $1`

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
		return existing, errors.WithStack(err)
	}

	if _, err = tx.Exec(selectForUpdate, definitionID); err != nil {
		return existing, errors.WithStack(err)
	}

	if _, err = tx.Exec(deletePorts, definitionID); err != nil {
		return existing, errors.WithStack(err)
	}

	if _, err = tx.Exec(deleteTags, definitionID); err != nil {
		return existing, errors.WithStack(err)
	}

	update := `
    UPDATE task_def SET
      image = $2,
      alias = $3,
      memory = $4,
      command = $5,
      env = $6,
      cpu = $7,
      gpu = $8,
      adaptive_resource_allocation = $9
    WHERE definition_id = $1;
    `
	if _, err = tx.Exec(
		update,
		definitionID,
		existing.Image,
		existing.Alias,
		existing.Memory,
		existing.Command,
		existing.Env,
		existing.Cpu,
		existing.Gpu,
		existing.AdaptiveResourceAllocation); err != nil {
		return existing, errors.Wrapf(err, "issue updating definition [%s]", definitionID)
	}

	if existing.Ports != nil {
		for _, p := range *existing.Ports {
			if _, err = tx.Exec(insertPorts, definitionID, p); err != nil {
				tx.Rollback()
				return existing, errors.WithStack(err)
			}
		}
	}

	if existing.Tags != nil {
		for _, t := range *existing.Tags {
			if _, err = tx.Exec(insertTags, t, t); err != nil {
				tx.Rollback()
				return existing, errors.WithStack(err)
			}
			if _, err = tx.Exec(insertDefTags, definitionID, t); err != nil {
				tx.Rollback()
				return existing, errors.WithStack(err)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return existing, errors.WithStack(err)
	}
	return existing, nil
}

// CreateDefinition creates the passed in definition object
// - error if definition already exists
func (sm *SQLStateManager) CreateDefinition(d Definition) error {
	var err error

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
		return errors.WithStack(err)
	}

	insert := `
    INSERT INTO task_def(
      definition_id,
      image,
      group_name,
      alias,
      memory,
      command,
      env,
      cpu,
      gpu,
      adaptive_resource_allocation
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
    `

	if _, err = tx.Exec(insert,
		d.DefinitionID,
		d.Image,
		d.GroupName,
		d.Alias,
		d.Memory,
		d.Command,
		d.Env,
		d.Cpu,
		d.Gpu,
		d.AdaptiveResourceAllocation); err != nil {
		tx.Rollback()
		return errors.Wrapf(
			err, "issue creating new task definition with alias [%s] and id [%s]", d.DefinitionID, d.Alias)
	}

	if d.Ports != nil {
		for _, p := range *d.Ports {
			if _, err = tx.Exec(insertPorts, d.DefinitionID, p); err != nil {
				tx.Rollback()
				return errors.WithStack(err)
			}
		}
	}

	if d.Tags != nil {
		for _, t := range *d.Tags {
			if _, err = tx.Exec(insertTags, t, t); err != nil {
				tx.Rollback()
				return errors.WithStack(err)
			}
			if _, err = tx.Exec(insertDefTags, d.DefinitionID, t); err != nil {
				tx.Rollback()
				return errors.WithStack(err)
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// DeleteDefinition deletes definition and associated runs and environment variables
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
		return errors.WithStack(err)
	}

	for _, stmt := range statements {
		if _, err = tx.Exec(stmt, definitionID); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "issue deleting definition with id [%s]", definitionID)
		}
	}

	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ListRuns returns a RunList
// limit: limit the result to this many runs
// offset: start the results at this offset
// sortBy: sort by this field
// order: 'asc' or 'desc'
// filters: map of field filters on Run - joined with AND
// envFilters: map of environment variable filters - joined with AND
func (sm *SQLStateManager) ListRuns(limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string, engines []string) (RunList, error) {

	var err error
	var result RunList
	var whereClause, orderQuery string

	if filters == nil {
		filters = make(map[string][]string)
	}

	if engines != nil {
		filters["engine"] = engines
	} else {
		filters["engine"] = []string{DefaultEngine}
	}

	where := append(sm.makeWhereClause(filters), sm.makeEnvWhereClause(envFilters)...)
	if len(where) > 0 {
		whereClause = fmt.Sprintf("where %s", strings.Join(where, " and "))
	}

	orderQuery, err = sm.orderBy(&Run{}, sortBy, order)
	if err != nil {
		return result, errors.WithStack(err)
	}

	sql := fmt.Sprintf(ListRunsSQL, whereClause, orderQuery)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Runs, sql, limit, offset)
	if err != nil {
		return result, errors.Wrap(err, "issue running list runs sql")
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list runs count sql")
	}

	return result, nil
}

// GetRun gets run by id
func (sm *SQLStateManager) GetRun(runID string) (Run, error) {
	var err error
	var r Run
	err = sm.db.Get(&r, GetRunSQL, runID)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, exceptions.MissingResource{
				fmt.Sprintf("Run with id %s not found", runID)}
		} else {
			return r, errors.Wrapf(err, "issue getting run with id [%s]", runID)
		}
	}
	return r, nil
}

func (sm *SQLStateManager) GetRunByEMRJobId(emrJobId string) (Run, error) {
	var err error
	var r Run
	err = sm.db.Get(&r, GetRunSQLByEMRJobId, emrJobId)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, exceptions.MissingResource{
				fmt.Sprintf("Run with emrjobid %s not found", emrJobId)}
		} else {
			return r, errors.Wrapf(err, "issue getting run with emrjobid [%s]", emrJobId)
		}
	}
	return r, nil
}

func (sm *SQLStateManager) GetResources(runID string) (Run, error) {
	var err error
	var r Run
	err = sm.db.Get(&r, GetRunSQL, runID)
	if err != nil {
		if err == sql.ErrNoRows {
			return r, exceptions.MissingResource{
				fmt.Sprintf("Run with id %s not found", runID)}
		} else {
			return r, errors.Wrapf(err, "issue getting run with id [%s]", runID)
		}
	}
	return r, nil
}

// UpdateRun updates run with updates - can be partial
func (sm *SQLStateManager) UpdateRun(runID string, updates Run) (Run, error) {
	start := time.Now()
	var (
		err      error
		existing Run
	)

	tx, err := sm.db.Begin()
	if err != nil {
		return existing, errors.WithStack(err)
	}

	rows, err := tx.Query(GetRunSQLForUpdate, runID)
	if err != nil {
		tx.Rollback()
		return existing, errors.WithStack(err)
	}

	for rows.Next() {
		err = rows.Scan(
			&existing.RunID,
			&existing.DefinitionID,
			&existing.Alias,
			&existing.Image,
			&existing.ClusterName,
			&existing.ExitCode,
			&existing.ExitReason,
			&existing.Status,
			&existing.QueuedAt,
			&existing.StartedAt,
			&existing.FinishedAt,
			&existing.InstanceID,
			&existing.InstanceDNSName,
			&existing.GroupName,
			&existing.TaskType,
			&existing.Env,
			&existing.Command,
			&existing.Memory,
			&existing.Cpu,
			&existing.Gpu,
			&existing.Engine,
			&existing.EphemeralStorage,
			&existing.NodeLifecycle,
			&existing.PodName,
			&existing.Namespace,
			&existing.MaxCpuUsed,
			&existing.MaxMemoryUsed,
			&existing.PodEvents,
			&existing.CommandHash,
			&existing.CloudTrailNotifications,
			&existing.ExecutableID,
			&existing.ExecutableType,
			&existing.ExecutionRequestCustom,
			&existing.CpuLimit,
			&existing.MemoryLimit,
			&existing.AttemptCount,
			&existing.SpawnedRuns,
			&existing.RunExceptions,
			&existing.ActiveDeadlineSeconds,
			&existing.SparkExtension,
			&existing.MetricsUri,
			&existing.Description,
			&existing.IdempotenceKey,
			&existing.User,
			&existing.Arch,
			&existing.Labels,
		)
	}
	if err != nil {
		return existing, errors.WithStack(err)
	}

	existing.UpdateWith(updates)

	update := `
    UPDATE task SET
        definition_id = $2,
		alias = $3,
		image = $4,
		cluster_name = $5,
		exit_code = $6,
		exit_reason = $7,
		status = $8,
		queued_at = $9,
		started_at = $10,
		finished_at = $11,
		instance_id = $12,
		instance_dns_name = $13,
		group_name = $14,
		env = $15,
		command = $16,
		memory = $17,
		cpu = $18,
		gpu = $19,
		engine = $20,
		ephemeral_storage = $21,
		node_lifecycle = $22,
		pod_name = $23,
		namespace = $24,
		max_cpu_used = $25,
		max_memory_used = $26,
		pod_events = $27,
		cloudtrail_notifications = $28,
		executable_id = $29,
		executable_type = $30,
		execution_request_custom = $31,
		cpu_limit = $32,
		memory_limit = $33,
		attempt_count = $34,
		spawned_runs = $35,
		run_exceptions = $36,
		active_deadline_seconds = $37,
		spark_extension = $38,
		metrics_uri = $39,
		description = $40,
		idempotence_key = $41,
		"user" = $42,
		arch = $43,
		labels = $44
    WHERE run_id = $1;
    `

	if _, err = tx.Exec(
		update,
		runID,
		existing.DefinitionID,
		existing.Alias,
		existing.Image,
		existing.ClusterName,
		existing.ExitCode,
		existing.ExitReason,
		existing.Status,
		existing.QueuedAt,
		existing.StartedAt,
		existing.FinishedAt,
		existing.InstanceID,
		existing.InstanceDNSName,
		existing.GroupName,
		existing.Env,
		existing.Command,
		existing.Memory,
		existing.Cpu,
		existing.Gpu,
		existing.Engine,
		existing.EphemeralStorage,
		existing.NodeLifecycle,
		existing.PodName,
		existing.Namespace,
		existing.MaxCpuUsed,
		existing.MaxMemoryUsed,
		existing.PodEvents,
		existing.CloudTrailNotifications,
		existing.ExecutableID,
		existing.ExecutableType,
		existing.ExecutionRequestCustom,
		existing.CpuLimit,
		existing.MemoryLimit,
		existing.AttemptCount,
		existing.SpawnedRuns,
		existing.RunExceptions,
		existing.ActiveDeadlineSeconds,
		existing.SparkExtension,
		existing.MetricsUri,
		existing.Description,
		existing.IdempotenceKey,
		existing.User,
		existing.Arch,
		existing.Labels); err != nil {
		tx.Rollback()
		return existing, errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		return existing, errors.WithStack(err)
	}

	_ = metrics.Timing(metrics.EngineUpdateRun, time.Since(start), []string{existing.ClusterName}, 1)

	return existing, nil
}

// CreateRun creates the passed in run
func (sm *SQLStateManager) CreateRun(r Run) error {
	var err error
	insert := `
	INSERT INTO task (
      	run_id,
		definition_id,
		alias,
		image,
		cluster_name,
		exit_code,
		exit_reason,
		status,
		queued_at,
		started_at,
		finished_at,
		instance_id,
		instance_dns_name,
		group_name,
		env,
		command,
		memory,
		cpu,
		gpu,
		engine,
		node_lifecycle,
		ephemeral_storage,
		pod_name,
		namespace,
		max_cpu_used,
		max_memory_used,
		pod_events,
		executable_id,
		executable_type,
		execution_request_custom,
		cpu_limit,
		memory_limit,
		attempt_count,
		spawned_runs,
		run_exceptions,
		active_deadline_seconds,
		task_type,
		command_hash,
		spark_extension,
		metrics_uri,
		description,
	    idempotence_key,
	    "user",
	    arch,
	    labels
    ) VALUES (
        $1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11,
		$12,
		$13,
		$14,
		$15,
		$16,
		$17,
		$18,
		$19,
		$20,
		$21,
		$22,
		$23,
		$24,
		$25,
		$26,
		$27,
		$28,
		$29,
		$30,
		$31,
		$32,
		$33,
		$34,
		$35,
		$36,
		$37,
		$38,
		$39,
		$40,
		$41,
        $42,
        $43,
        $44,
        $45
	);
    `

	tx, err := sm.db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}

	if _, err = tx.Exec(insert,
		r.RunID,
		r.DefinitionID,
		r.Alias,
		r.Image,
		r.ClusterName,
		r.ExitCode,
		r.ExitReason,
		r.Status,
		r.QueuedAt,
		r.StartedAt,
		r.FinishedAt,
		r.InstanceID,
		r.InstanceDNSName,
		r.GroupName,
		r.Env,
		r.Command,
		r.Memory,
		r.Cpu,
		r.Gpu,
		r.Engine,
		r.NodeLifecycle,
		r.EphemeralStorage,
		r.PodName,
		r.Namespace,
		r.MaxCpuUsed,
		r.MaxMemoryUsed,
		r.PodEvents,
		r.ExecutableID,
		r.ExecutableType,
		r.ExecutionRequestCustom,
		r.CpuLimit,
		r.MemoryLimit,
		r.AttemptCount,
		r.SpawnedRuns,
		r.RunExceptions,
		r.ActiveDeadlineSeconds,
		r.TaskType,
		r.CommandHash,
		r.SparkExtension,
		r.MetricsUri,
		r.Description,
		r.IdempotenceKey,
		r.User,
		r.Arch,
		r.Labels); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "issue creating new task run with id [%s]", r.RunID)
	}

	if err = tx.Commit(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// ListGroups returns a list of the existing group names.
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
		return result, errors.Wrap(err, "issue running list groups sql")
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list groups count sql")
	}

	return result, nil
}

// ListTags returns a list of the existing tags.
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
		return result, errors.Wrap(err, "issue running list tags sql")
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list tags count sql")
	}

	return result, nil
}

// initWorkerTable initializes the `worker` table with values from the config
func (sm *SQLStateManager) initWorkerTable(c config.Config) error {
	// Get worker count from configuration (set to 1 as default)

	for _, engine := range Engines {
		retryCount := int64(1)
		if c.IsSet(fmt.Sprintf("worker.%s.retry_worker_count_per_instance", engine)) {
			retryCount = int64(c.GetInt("worker.ecs.retry_worker_count_per_instance"))
		}
		submitCount := int64(1)
		if c.IsSet(fmt.Sprintf("worker.%s.submit_worker_count_per_instance", engine)) {
			submitCount = int64(c.GetInt("worker.ecs.submit_worker_count_per_instance"))
		}
		statusCount := int64(1)
		if c.IsSet(fmt.Sprintf("worker.%s.status_worker_count_per_instance", engine)) {
			statusCount = int64(c.GetInt("worker.ecs.status_worker_count_per_instance"))
		}

		var err error
		insert := `
		INSERT INTO worker (worker_type, count_per_instance, engine)
		VALUES ('retry', $1, $4), ('submit', $2, $4), ('status', $3, $4);
	`

		tx, err := sm.db.Begin()
		if err != nil {
			return errors.WithStack(err)
		}

		if _, err = tx.Exec(insert, retryCount, submitCount, statusCount, engine); err != nil {
			tx.Rollback()
			return errors.Wrapf(err, "issue populating worker table")
		}

		err = tx.Commit()

		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

// ListWorkers returns list of workers
func (sm *SQLStateManager) ListWorkers(engine string) (WorkersList, error) {
	var err error
	var result WorkersList

	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", ListWorkersSQL)

	err = sm.readonlyDB.Select(&result.Workers, GetWorkerEngine, engine)
	if err != nil {
		return result, errors.Wrap(err, "issue running list workers sql")
	}

	err = sm.readonlyDB.Get(&result.Total, countSQL)
	if err != nil {
		return result, errors.Wrap(err, "issue running list workers count sql")
	}

	return result, nil
}

// GetWorker returns data for a single worker.
func (sm *SQLStateManager) GetWorker(workerType string, engine string) (w Worker, err error) {
	if err := sm.readonlyDB.Get(&w, GetWorkerSQL, workerType, engine); err != nil {
		if err == sql.ErrNoRows {
			err = exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Worker of type %s not found", workerType)}
		} else {
			err = errors.Wrapf(err, "issue getting worker of type [%s]", workerType)
		}
	}
	return
}

// UpdateWorker updates a single worker.
func (sm *SQLStateManager) UpdateWorker(workerType string, updates Worker) (Worker, error) {
	var (
		err      error
		existing Worker
	)

	engine := DefaultEngine
	tx, err := sm.db.Begin()
	if err != nil {
		return existing, errors.WithStack(err)
	}

	rows, err := tx.Query(GetWorkerSQLForUpdate, workerType, engine)
	if err != nil {
		tx.Rollback()
		return existing, errors.WithStack(err)
	}

	for rows.Next() {
		err = rows.Scan(&existing.WorkerType, &existing.CountPerInstance)
	}
	if err != nil {
		return existing, errors.WithStack(err)
	}

	existing.UpdateWith(updates)

	update := `
		UPDATE worker SET count_per_instance = $2
    WHERE worker_type = $1;
    `

	if _, err = tx.Exec(update, workerType, existing.CountPerInstance); err != nil {
		tx.Rollback()
		return existing, errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		return existing, errors.WithStack(err)
	}

	return existing, nil
}

// BatchUpdateWorker updates multiple workers.
func (sm *SQLStateManager) BatchUpdateWorkers(updates []Worker) (WorkersList, error) {
	var existing WorkersList

	for _, w := range updates {
		_, err := sm.UpdateWorker(w.WorkerType, w)

		if err != nil {
			return existing, err
		}
	}

	return sm.ListWorkers(DefaultEngine)
}

// Cleanup close any open resources
func (sm *SQLStateManager) Cleanup() error {
	return multierr.Combine(sm.db.Close(), sm.readonlyDB.Close())
}

type IOrderable interface {
	ValidOrderField(field string) bool
	ValidOrderFields() []string
	DefaultOrderField() string
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

func (d *Definition) DefaultOrderField() string {
	return "group_name"
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

func (r *Run) DefaultOrderField() string {
	return "group_name"
}

func (t *Template) ValidOrderField(field string) bool {
	for _, f := range t.ValidOrderFields() {
		if field == f {
			return true
		}
	}
	return false
}

func (t *Template) ValidOrderFields() []string {
	// @TODO: figure what fields should be orderable.
	return []string{"template_name", "version"}
}

func (t *Template) DefaultOrderField() string {
	return "template_name"
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
func (e *PodEvents) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e SpawnedRuns) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

func (e *SpawnedRuns) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e SparkExtension) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

func (e *SparkExtension) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e RunExceptions) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

func (e *RunExceptions) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e PodEvents) Value() (driver.Value, error) {
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

// Scan from db
func (e *CloudTrailNotifications) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e CloudTrailNotifications) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

// Scan from db
func (e *ExecutionRequestCustom) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// Value to db
func (e ExecutionRequestCustom) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

// Scan from db
func (tjs *TemplateJSONSchema) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.([]uint8))
		json.Unmarshal(s, &tjs)
	}
	return nil
}

// Value to db
func (tjs TemplateJSONSchema) Value() (driver.Value, error) {
	res, _ := json.Marshal(tjs)
	return res, nil
}

// Scan from db
func (tjs *TemplatePayload) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.([]uint8))
		json.Unmarshal(s, &tjs)
	}
	return nil
}

// Value to db
func (tjs TemplatePayload) Value() (driver.Value, error) {
	res, _ := json.Marshal(tjs)
	return res, nil
}

// Value to db
func (e Labels) Value() (driver.Value, error) {
	res, _ := json.Marshal(e)
	return res, nil
}

func (e *Labels) Scan(value interface{}) error {
	if value != nil {
		s := []byte(value.(string))
		json.Unmarshal(s, &e)
	}
	return nil
}

// GetTemplateByID returns a single template by id.
func (sm *SQLStateManager) GetTemplateByID(templateID string) (Template, error) {
	var err error
	var tpl Template
	err = sm.db.Get(&tpl, GetTemplateByIDSQL, templateID)
	if err != nil {
		if err == sql.ErrNoRows {
			return tpl, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Template with ID %s not found", templateID)}
		}

		return tpl, errors.Wrapf(err, "issue getting tpl with id [%s]", templateID)
	}
	return tpl, nil
}

func (sm *SQLStateManager) GetTemplateByVersion(templateName string, templateVersion int64) (bool, Template, error) {
	var err error
	var tpl Template
	err = sm.db.Get(&tpl, GetTemplateByVersionSQL, templateName, templateVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, tpl, nil
		}

		return false, tpl, errors.Wrapf(err, "issue getting tpl with id [%s]", templateName)
	}
	return true, tpl, nil
}

// GetLatestTemplateByTemplateName returns the latest version of a template
// of a specific template name.
func (sm *SQLStateManager) GetLatestTemplateByTemplateName(templateName string) (bool, Template, error) {
	var err error
	var tpl Template
	err = sm.db.Get(&tpl, GetTemplateLatestOnlySQL, templateName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, tpl, nil
		}

		return false, tpl, errors.Wrapf(err, "issue getting tpl with id [%s]", templateName)
	}
	return true, tpl, nil
}

// ListTemplates returns list of templates from the database.
func (sm *SQLStateManager) ListTemplates(limit int, offset int, sortBy string, order string) (TemplateList, error) {
	var err error
	var result TemplateList
	var orderQuery string

	orderQuery, err = sm.orderBy(&Template{}, sortBy, order)
	if err != nil {
		return result, errors.WithStack(err)
	}

	sql := fmt.Sprintf(ListTemplatesSQL, orderQuery)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.Select(&result.Templates, sql, limit, offset)
	if err != nil {
		return result, errors.Wrap(err, "issue running list templates sql")
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list templates count sql")
	}

	return result, nil
}

// ListTemplates returns list of templates from the database.
func (sm *SQLStateManager) ListTemplatesLatestOnly(limit int, offset int, sortBy string, order string) (TemplateList, error) {
	var err error
	var result TemplateList

	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", ListTemplatesLatestOnlySQL)

	err = sm.db.Select(&result.Templates, ListTemplatesLatestOnlySQL, limit, offset)
	if err != nil {
		return result, errors.Wrap(err, "issue running list templates sql")
	}
	err = sm.db.Get(&result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list templates count sql")
	}

	return result, nil
}

// CreateTemplate creates a new template.
func (sm *SQLStateManager) CreateTemplate(t Template) error {
	var err error
	insert := `
    INSERT INTO template(
			template_id, template_name, version, schema, command_template,
			adaptive_resource_allocation, image, memory, env, cpu, gpu, defaults, avatar_uri
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);
    `

	tx, err := sm.db.Begin()
	if err != nil {
		return errors.WithStack(err)
	}

	if _, err = tx.Exec(insert,
		t.TemplateID, t.TemplateName, t.Version, t.Schema, t.CommandTemplate,
		t.AdaptiveResourceAllocation, t.Image, t.Memory, t.Env,
		t.Cpu, t.Gpu, t.Defaults, t.AvatarURI); err != nil {
		tx.Rollback()
		return errors.Wrapf(
			err, "issue creating new template with template_name [%s] and version [%d]", t.TemplateName, t.Version)
	}

	err = tx.Commit()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// GetExecutableByExecutableType returns a single executable by id.
func (sm *SQLStateManager) GetExecutableByTypeAndID(t ExecutableType, id string) (Executable, error) {
	switch t {
	case ExecutableTypeDefinition:
		return sm.GetDefinition(id)
	case ExecutableTypeTemplate:
		return sm.GetTemplateByID(id)
	default:
		return nil, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("executable type of [%s] not valid.", t),
		}
	}
}
