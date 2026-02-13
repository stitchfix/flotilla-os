package state

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/stitchfix/flotilla-os/clients/metrics"
	"github.com/stitchfix/flotilla-os/log"
	"github.com/stitchfix/flotilla-os/tracing"

	"github.com/jmoiron/sqlx"

	// Pull in postgres specific drivers
	"database/sql"
	"math"
	"strings"

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
	log        log.Logger
}

func (sm *SQLStateManager) ListFailingNodes(ctx context.Context) (NodeList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_failing_nodes", "")
	defer span.Finish()

	var err error
	var nodeList NodeList

	err = sm.readonlyDB.SelectContext(ctx, &nodeList, ListFailingNodesSQL)

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

func (sm *SQLStateManager) GetPodReAttemptRate(ctx context.Context) (float32, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_pod_reattempt_rate", "")
	defer span.Finish()

	var err error
	attemptRate := float32(1.0)
	err = sm.readonlyDB.GetContext(ctx, &attemptRate, PodReAttemptRate)

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

func (sm *SQLStateManager) GetNodeLifecycle(ctx context.Context, executableID string, commandHash string) (string, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_node_lifecycle", "")
	defer span.Finish()
	//span.SetTag("command_hash", commandHash)

	var err error
	nodeType := "spot"
	err = sm.readonlyDB.GetContext(ctx, &nodeType, TaskResourcesExecutorNodeLifecycleSQL, executableID, commandHash)

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

func (sm *SQLStateManager) GetTaskHistoricalRuntime(ctx context.Context, executableID string, runID string) (float32, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_task_historical_runtime", "")
	defer span.Finish()

	span.SetTag("job.run_id", runID)

	var err error
	minutes := float32(1.0)
	err = sm.readonlyDB.GetContext(ctx, &minutes, TaskExecutionRuntimeCommandSQL, executableID, runID)

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

func (sm *SQLStateManager) EstimateRunResources(ctx context.Context, executableID string, commandHash string) (TaskResources, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.estimate_run_resources", "")
	defer span.Finish()

	//span.SetTag("command_hash", commandHash)

	var err error
	var taskResources TaskResources

	err = sm.readonlyDB.GetContext(ctx, &taskResources, TaskResourcesSelectCommandSQL, executableID, commandHash)

	if err != nil {
		if err == sql.ErrNoRows {
			// No historical data found - this is expected for new jobs or jobs that haven't OOM'd
			if sm.log != nil {
				_ = sm.log.Log(
					"level", "info",
					"message", "ARA: No historical resource data found",
					"definition_id", executableID,
					"command_hash", commandHash,
				)
			}
			return taskResources, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Resource usage with executable %s not found", executableID)}
		} else {
			// Check if this is a PostgreSQL recovery conflict (expected on read replicas)
			errMsg := err.Error()
			isRecoveryConflict := strings.Contains(errMsg, "conflict with recovery") ||
				strings.Contains(errMsg, "canceling statement due to conflict")

			if isRecoveryConflict {
				// Recovery conflicts are expected on read replicas - treat as missing data
				// Log at info level since this is expected behavior, not an error
				if sm.log != nil {
					_ = sm.log.Log(
						"level", "info",
						"message", "ARA: Query canceled due to recovery conflict on read replica (using defaults)",
						"definition_id", executableID,
						"command_hash", commandHash,
					)
				}
				return taskResources, exceptions.MissingResource{
					ErrorString: fmt.Sprintf("Resource usage with executable %s not found (recovery conflict)", executableID)}
			}

			// Unexpected error querying historical data
			if sm.log != nil {
				_ = sm.log.Log(
					"level", "error",
					"message", "ARA: Error querying historical resource data",
					"definition_id", executableID,
					"command_hash", commandHash,
					"error", err.Error(),
				)
			}
			return taskResources, errors.Wrapf(err, "issue getting resources with executable [%s]", executableID)
		}
	}

	// Check if the query returned NULL values (can happen when percentile_disc has no valid data)
	if !taskResources.Memory.Valid || !taskResources.Cpu.Valid {
		// NULL values mean no valid historical data - treat as missing resource
		if sm.log != nil {
			_ = sm.log.Log(
				"level", "info",
				"message", "ARA: No historical resource data found (NULL values returned)",
				"definition_id", executableID,
				"command_hash", commandHash,
			)
		}
		return taskResources, exceptions.MissingResource{
			ErrorString: fmt.Sprintf("Resource usage with executable %s not found (NULL values)", executableID)}
	}

	// Successfully found historical data - log the values being returned
	if sm.log != nil {
		_ = sm.log.Log(
			"level", "info",
			"message", "ARA: Historical resource data found",
			"definition_id", executableID,
			"command_hash", commandHash,
			"estimated_memory_mb", taskResources.Memory.Int64,
			"estimated_cpu_millicores", taskResources.Cpu.Int64,
		)
	}

	return taskResources, err
}

func (sm *SQLStateManager) EstimateExecutorCount(ctx context.Context, executableID string, commandHash string) (int64, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.estimate_executor_count", "")
	defer span.Finish()

	//span.SetTag("command_hash", commandHash)

	var err error
	executorCount := int64(25)
	err = sm.readonlyDB.GetContext(ctx, &executorCount, TaskResourcesExecutorCountSQL, executableID, commandHash)

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
func (sm *SQLStateManager) CheckIdempotenceKey(ctx context.Context, idempotenceKey string) (string, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.check_idempotence_key", "")
	defer span.Finish()

	var err error
	runId := ""
	err = sm.readonlyDB.GetContext(ctx, &runId, TaskIdempotenceKeyCheckSQL, idempotenceKey)

	if err != nil || len(runId) == 0 {
		err = errors.New("no run_id found for idempotence key")
	}
	return runId, err
}

func (sm *SQLStateManager) ExecutorOOM(ctx context.Context, executableID string, commandHash string) (bool, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.executor_oom", "")
	defer span.Finish()

	//span.SetTag("command_hash", commandHash)

	var err error
	executorOOM := false
	err = sm.readonlyDB.GetContext(ctx, &executorOOM, TaskResourcesExecutorOOMSQL, executableID, commandHash)

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

func (sm *SQLStateManager) DriverOOM(ctx context.Context, executableID string, commandHash string) (bool, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.driver_oom", "")
	defer span.Finish()

	//span.SetTag("command_hash", commandHash)

	var err error
	driverOOM := false
	err = sm.readonlyDB.GetContext(ctx, &driverOOM, TaskResourcesDriverOOMSQL, executableID, commandHash)

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
	fmt.Printf("create_database_schema: %t\ncreating schema...\n", createSchema)
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
	ctx context.Context,
	limit int, offset int, sortBy string,
	order string, filters map[string][]string,
	envFilters map[string]string) (DefinitionList, error) {
	// Use "list" as an identifier since there's no specific runID for a list operation
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_definitions", "")
	defer span.Finish()

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
func (sm *SQLStateManager) GetDefinition(ctx context.Context, definitionID string) (Definition, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_definition", "")
	defer span.Finish()

	var err error
	var definition Definition
	err = sm.db.GetContext(ctx, &definition, GetDefinitionSQL, definitionID)
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
func (sm *SQLStateManager) GetDefinitionByAlias(ctx context.Context, alias string) (Definition, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_definition_by_alias", "")
	defer span.Finish()

	//span.SetTag("alias", alias)

	var err error
	var definition Definition
	err = sm.db.GetContext(ctx, &definition, GetDefinitionByAliasSQL, alias)
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
func (sm *SQLStateManager) UpdateDefinition(ctx context.Context, definitionID string, updates Definition) (Definition, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.update_definition", "")
	defer span.Finish()
	var (
		err      error
		existing Definition
	)
	existing, err = sm.GetDefinition(ctx, definitionID)
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
      adaptive_resource_allocation = $9,
      ephemeral_storage = $10,
	  requires_docker = $11,
      target_cluster = $12
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
		existing.AdaptiveResourceAllocation,
		existing.EphemeralStorage,
		existing.RequiresDocker,
		existing.TargetCluster); err != nil {
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
func (sm *SQLStateManager) CreateDefinition(ctx context.Context, d Definition) error {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.create_definition", "")
	defer span.Finish()
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
      adaptive_resource_allocation,
      ephemeral_storage,
      requires_docker,
      target_cluster
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);
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
		d.AdaptiveResourceAllocation,
		d.EphemeralStorage,
		d.RequiresDocker,
		d.TargetCluster); err != nil {
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
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return errors.WithStack(err)
	}
	return nil
}

// DeleteDefinition deletes definition and associated runs and environment variables
func (sm *SQLStateManager) DeleteDefinition(ctx context.Context, definitionID string) error {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.delete_definition", "")
	defer span.Finish()
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
func (sm *SQLStateManager) ListRuns(ctx context.Context, limit int, offset int, sortBy string, order string, filters map[string][]string, envFilters map[string]string, engines []string) (RunList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_runs", "")
	defer span.Finish()
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
func (sm *SQLStateManager) GetRun(ctx context.Context, runID string) (Run, error) {
	// Create a span for this database operation using the utils.TraceJob function
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_run", "")
	defer span.Finish()
	span.SetTag("job.run_id", runID)
	var r Run
	err := sm.db.GetContext(ctx, &r, GetRunSQL, runID)
	if err != nil {
		// Tag error for easier debugging
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())

		if err == sql.ErrNoRows {
			return r, exceptions.MissingResource{
				fmt.Sprintf("Run with id %s not found", runID)}
		} else {
			return r, errors.Wrapf(err, "issue getting run with id [%s]", runID)
		}
	}

	// Tag the span with run metadata
	tracing.TagRunInfo(span,
		r.RunID, r.DefinitionID, r.Alias, r.Status, r.ClusterName,
		r.QueuedAt, r.StartedAt, r.FinishedAt,
		r.PodName, r.Namespace, r.ExitReason, r.ExitCode, string(r.Tier))

	return r, nil
}

func (sm *SQLStateManager) GetRunByEMRJobId(ctx context.Context, emrJobId string) (Run, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_run_by_emr_job_id", "")
	defer span.Finish()
	span.SetTag("job.emr_job_id", emrJobId)
	var err error
	var r Run
	err = sm.db.GetContext(ctx, &r, GetRunSQLByEMRJobId, emrJobId)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		if err == sql.ErrNoRows {
			return r, exceptions.MissingResource{
				fmt.Sprintf("Run with emrjobid %s not found", emrJobId)}
		} else {
			return r, errors.Wrapf(err, "issue getting run with emrjobid [%s]", emrJobId)
		}
	}

	// Tag the span with run metadata
	tracing.TagRunInfo(span,
		r.RunID, r.DefinitionID, r.Alias, r.Status, r.ClusterName,
		r.QueuedAt, r.StartedAt, r.FinishedAt,
		r.PodName, r.Namespace, r.ExitReason, r.ExitCode, string(r.Tier))

	return r, nil
}

func (sm *SQLStateManager) GetResources(ctx context.Context, runID string) (Run, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_resources", "")
	defer span.Finish()
	span.SetTag("job.run_id", runID)
	var err error
	var r Run
	err = sm.db.GetContext(ctx, &r, GetRunSQL, runID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		if err == sql.ErrNoRows {
			return r, exceptions.MissingResource{
				fmt.Sprintf("Run with id %s not found", runID)}
		} else {
			return r, errors.Wrapf(err, "issue getting run with id [%s]", runID)
		}
	}

	// Tag the span with run metadata
	tracing.TagRunInfo(span,
		r.RunID, r.DefinitionID, r.Alias, r.Status, r.ClusterName,
		r.QueuedAt, r.StartedAt, r.FinishedAt,
		r.PodName, r.Namespace, r.ExitReason, r.ExitCode, string(r.Tier))

	return r, nil
}

// UpdateRun updates run with updates - can be partial
func (sm *SQLStateManager) UpdateRun(ctx context.Context, runID string, updates Run) (Run, error) {
	start := time.Now()
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.update_run", "")
	defer span.Finish()
	span.SetTag("job.run_id", runID)
	span.SetTag("status", updates.Status)
	var (
		err      error
		existing Run
	)

	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("error.type", "begin_transaction")
		return existing, errors.WithStack(err)
	}

	rows, err := tx.QueryContext(ctx, GetRunSQLForUpdate, runID)
	if err != nil {
		tx.Rollback()
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		span.SetTag("error.type", "query")
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
			&existing.RequiresDocker,
			&existing.ServiceAccount,
			&existing.Tier,
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
		labels = $44,
		requires_docker = $45,
		service_account = $46,
        tier = $47
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
		existing.Labels,
		existing.RequiresDocker,
		existing.ServiceAccount,
		existing.Tier); err != nil {
		tx.Rollback()
		return existing, errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		return existing, errors.WithStack(err)
	}

	_ = metrics.Timing(metrics.EngineUpdateRun, time.Since(start), []string{existing.ClusterName}, 1)
	go sm.logStatusUpdate(existing)
	return existing, nil
}

// CreateRun creates the passed in run
func (sm *SQLStateManager) CreateRun(ctx context.Context, r Run) error {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.create_run", "")
	defer span.Finish()
	span.SetTag("job.run_id", r.RunID)
	// Now utils.TraceJob already sets the run_id tag
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
	    labels,
		requires_docker,
		service_account,
		tier
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
        $45,
    	$46,
    	$47,
    	$48
	);
    `

	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return errors.WithStack(err)
	}

	if _, err = tx.ExecContext(ctx, insert,
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
		r.Labels,
		r.RequiresDocker,
		r.ServiceAccount,
		r.Tier); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "issue creating new task run with id [%s]", r.RunID)
	}

	if err = tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	go sm.logStatusUpdate(r)
	return nil
}

// ListGroups returns a list of the existing group names.
func (sm *SQLStateManager) ListGroups(ctx context.Context, limit int, offset int, name *string) (GroupsList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_groups", "")
	defer span.Finish()
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
func (sm *SQLStateManager) ListTags(ctx context.Context, limit int, offset int, name *string) (TagsList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_tags", "")
	defer span.Finish()
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

	err = sm.db.SelectContext(ctx, &result.Tags, sql, limit, offset)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return result, errors.Wrap(err, "issue running list tags sql")
	}
	err = sm.db.GetContext(ctx, &result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list tags count sql")
	}

	return result, nil
}

// initWorkerTable initializes the `worker` table with values from the config
func (sm *SQLStateManager) initWorkerTable(c config.Config) error {
	// Get worker count from configuration (set to 1 as default)

	for _, engine := range Engines {
		fmt.Printf("init worker table for %s engine", engine)
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
func (sm *SQLStateManager) ListWorkers(ctx context.Context, engine string) (WorkersList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_workers", "")
	defer span.Finish()
	var err error
	var result WorkersList

	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", ListWorkersSQL)

	err = sm.readonlyDB.SelectContext(ctx, &result.Workers, GetWorkerEngine, engine)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return result, errors.Wrap(err, "issue running list workers sql")
	}

	err = sm.readonlyDB.GetContext(ctx, &result.Total, countSQL)
	if err != nil {
		return result, errors.Wrap(err, "issue running list workers count sql")
	}

	return result, nil
}

// GetWorker returns data for a single worker.
func (sm *SQLStateManager) GetWorker(ctx context.Context, workerType string, engine string) (w Worker, err error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_worker", "")
	defer span.Finish()
	//span.SetTag("engine", engine)
	if err := sm.readonlyDB.GetContext(ctx, &w, GetWorkerSQL, workerType, engine); err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
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
func (sm *SQLStateManager) UpdateWorker(ctx context.Context, workerType string, updates Worker) (Worker, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.update_worker", "")
	defer span.Finish()
	var (
		err      error
		existing Worker
	)

	engine := DefaultEngine
	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return existing, errors.WithStack(err)
	}

	rows, err := tx.QueryContext(ctx, GetWorkerSQLForUpdate, workerType, engine)
	if err != nil {
		tx.Rollback()
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return existing, errors.WithStack(err)
	}

	for rows.Next() {
		err = rows.Scan(&existing.WorkerType, &existing.CountPerInstance, &existing.Engine)
	}
	if err != nil {
		return existing, errors.WithStack(err)
	}

	existing.UpdateWith(updates)

	update := `
		UPDATE worker SET count_per_instance = $2
    WHERE worker_type = $1;
    `

	if _, err = tx.ExecContext(ctx, update, workerType, existing.CountPerInstance); err != nil {
		tx.Rollback()
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return existing, errors.WithStack(err)
	}

	if err = tx.Commit(); err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return existing, errors.WithStack(err)
	}

	return existing, nil
}

// BatchUpdateWorker updates multiple workers.
func (sm *SQLStateManager) BatchUpdateWorkers(ctx context.Context, updates []Worker) (WorkersList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.batch_update_workers", "")
	defer span.Finish()
	var existing WorkersList

	for _, w := range updates {
		_, err := sm.UpdateWorker(ctx, w.WorkerType, w)

		if err != nil {
			span.SetTag("error", true)
			span.SetTag("error.msg", err.Error())
			return existing, err
		}
	}

	return sm.ListWorkers(ctx, DefaultEngine)
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
func (e *EnvList) Value() (driver.Value, error) {
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
func (sm *SQLStateManager) GetTemplateByID(ctx context.Context, templateID string) (Template, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_template_by_id", "")
	defer span.Finish()
	var err error
	var tpl Template
	err = sm.db.GetContext(ctx, &tpl, GetTemplateByIDSQL, templateID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		if err == sql.ErrNoRows {
			return tpl, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Template with ID %s not found", templateID)}
		}

		return tpl, errors.Wrapf(err, "issue getting tpl with id [%s]", templateID)
	}
	return tpl, nil
}

func (sm *SQLStateManager) GetTemplateByVersion(ctx context.Context, templateName string, templateVersion int64) (bool, Template, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_template_by_version", "")
	defer span.Finish()
	span.SetTag("template.version", templateVersion)
	var err error
	var tpl Template
	err = sm.db.GetContext(ctx, &tpl, GetTemplateByVersionSQL, templateName, templateVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, tpl, nil
		}

		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return false, tpl, errors.Wrapf(err, "issue getting tpl with id [%s]", templateName)
	}
	return true, tpl, nil
}

// GetLatestTemplateByTemplateName returns the latest version of a template
// of a specific template name.
func (sm *SQLStateManager) GetLatestTemplateByTemplateName(ctx context.Context, templateName string) (bool, Template, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_latest_template_by_name", "")
	defer span.Finish()
	var err error
	var tpl Template
	err = sm.db.GetContext(ctx, &tpl, GetTemplateLatestOnlySQL, templateName)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, tpl, nil
		}

		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return false, tpl, errors.Wrapf(err, "issue getting tpl with id [%s]", templateName)
	}
	return true, tpl, nil
}

// ListTemplates returns list of templates from the database.
func (sm *SQLStateManager) ListTemplates(ctx context.Context, limit int, offset int, sortBy string, order string) (TemplateList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_templates", "")
	defer span.Finish()
	var err error
	var result TemplateList
	var orderQuery string

	orderQuery, err = sm.orderBy(&Template{}, sortBy, order)
	if err != nil {
		return result, errors.WithStack(err)
	}

	sql := fmt.Sprintf(ListTemplatesSQL, orderQuery)
	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", sql)

	err = sm.db.SelectContext(ctx, &result.Templates, sql, limit, offset)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return result, errors.Wrap(err, "issue running list templates sql")
	}
	err = sm.db.GetContext(ctx, &result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list templates count sql")
	}

	return result, nil
}

// ListTemplatesLatestOnly returns list of templates from the database.
func (sm *SQLStateManager) ListTemplatesLatestOnly(ctx context.Context, limit int, offset int, sortBy string, order string) (TemplateList, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_templates_latest_only", "")
	defer span.Finish()
	var err error
	var result TemplateList

	countSQL := fmt.Sprintf("select COUNT(*) from (%s) as sq", ListTemplatesLatestOnlySQL)

	err = sm.db.SelectContext(ctx, &result.Templates, ListTemplatesLatestOnlySQL, limit, offset)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return result, errors.Wrap(err, "issue running list templates sql")
	}
	err = sm.db.GetContext(ctx, &result.Total, countSQL, nil, 0)
	if err != nil {
		return result, errors.Wrap(err, "issue running list templates count sql")
	}

	return result, nil
}

// CreateTemplate creates a new template.
func (sm *SQLStateManager) CreateTemplate(ctx context.Context, t Template) error {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.create_template", "")
	defer span.Finish()
	var err error
	insert := `
    INSERT INTO template(
			template_id, template_name, version, schema, command_template,
			adaptive_resource_allocation, image, memory, env, cpu, gpu, defaults, avatar_uri
    )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);
    `

	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return errors.WithStack(err)
	}

	if _, err = tx.ExecContext(ctx, insert,
		t.TemplateID, t.TemplateName, t.Version, t.Schema, t.CommandTemplate,
		t.AdaptiveResourceAllocation, t.Image, t.Memory, t.Env,
		t.Cpu, t.Gpu, t.Defaults, t.AvatarURI); err != nil {
		tx.Rollback()
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return errors.Wrapf(
			err, "issue creating new template with template_name [%s] and version [%d]", t.TemplateName, t.Version)
	}

	err = tx.Commit()
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return errors.WithStack(err)
	}
	return nil
}

// GetExecutableByExecutableType returns a single executable by id.
func (sm *SQLStateManager) GetExecutableByTypeAndID(ctx context.Context, t ExecutableType, id string) (Executable, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_executable_by_type_and_id", "")
	defer span.Finish()
	span.SetTag("executable.type", string(t))

	switch t {
	case ExecutableTypeDefinition:
		return sm.GetDefinition(ctx, id)
	case ExecutableTypeTemplate:
		return sm.GetTemplateByID(ctx, id)
	default:
		span.SetTag("error", true)
		span.SetTag("error.msg", fmt.Sprintf("executable type of [%s] not valid", t))
		return nil, exceptions.MalformedInput{
			ErrorString: fmt.Sprintf("executable type of [%s] not valid.", t),
		}
	}
}

func (sm *SQLStateManager) logStatusUpdate(update Run) {
	var err error
	var startedAt, finishedAt time.Time
	var duration float64
	var env EnvList
	var command string

	if update.StartedAt != nil {
		startedAt = *update.StartedAt
		duration = time.Now().Sub(startedAt).Seconds()
	}

	if update.FinishedAt != nil {
		finishedAt = *update.FinishedAt
		duration = finishedAt.Sub(startedAt).Seconds()
	}

	if update.Env != nil {
		env = *update.Env
	}

	if update.Command != nil {
		command = *update.Command
	}

	if update.ExitCode != nil {
		err = sm.log.Event("eventClassName", "FlotillaTaskStatus",
			"run_id", update.RunID,
			"definition_id", update.DefinitionID,
			"alias", update.Alias,
			"image", update.Image,
			"cluster_name", update.ClusterName,
			"command", command,
			"exit_code", *update.ExitCode,
			"status", update.Status,
			"started_at", startedAt,
			"finished_at", finishedAt,
			"duration", duration,
			"instance_id", update.InstanceID,
			"instance_dns_name", update.InstanceDNSName,
			"group_name", update.GroupName,
			"user", update.User,
			"task_type", update.TaskType,
			"env", env,
			"executable_id", update.ExecutableID,
			"executable_type", update.ExecutableType,
			"Tier", update.Tier)
	} else {
		err = sm.log.Event("eventClassName", "FlotillaTaskStatus",
			"run_id", update.RunID,
			"definition_id", update.DefinitionID,
			"alias", update.Alias,
			"image", update.Image,
			"cluster_name", update.ClusterName,
			"command", command,
			"status", update.Status,
			"started_at", startedAt,
			"finished_at", finishedAt,
			"duration", duration,
			"instance_id", update.InstanceID,
			"instance_dns_name", update.InstanceDNSName,
			"group_name", update.GroupName,
			"user", update.User,
			"task_type", update.TaskType,
			"env", env,
			"executable_id", update.ExecutableID,
			"executable_type", update.ExecutableType,
			"Tier", update.Tier)
	}

	if err != nil {
		sm.log.Log("level", "error", "message", "Failed to emit status event", "run_id", update.RunID, "error", err.Error())
	}
}

func (sm *SQLStateManager) ListClusterStates(ctx context.Context) ([]ClusterMetadata, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.list_cluster_states", "")
	defer span.Finish()

	var clusters []ClusterMetadata
	err := sm.db.SelectContext(ctx, &clusters, ListClusterStatesSQL)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
	}
	return clusters, err
}

func (sm *SQLStateManager) UpdateClusterMetadata(ctx context.Context, cluster ClusterMetadata) error {
	operationName := "flotilla.state.create_cluster_metadata"
	identifier := cluster.Name

	if cluster.ID != "" {
		operationName = "flotilla.state.update_cluster_metadata"
		identifier = cluster.ID
	}

	ctx, span := tracing.TraceJob(ctx, operationName, "")
	defer span.Finish()
	span.SetTag("cluster.id", identifier)
	// Add relevant tags
	span.SetTag("cluster.name", cluster.Name)
	span.SetTag("cluster.status", cluster.Status)
	if cluster.ClusterVersion != "" {
		span.SetTag("cluster.version", cluster.ClusterVersion)
	}

	if cluster.ID == "" {
		sql := `
			INSERT INTO cluster_state (name, cluster_version, status, status_reason, allowed_tiers, capabilities, namespace, region, emr_virtual_cluster, spark_server_uri)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id;
		`
		var id string
		err := sm.db.QueryRowContext(ctx, sql,
			cluster.Name,
			cluster.ClusterVersion,
			cluster.Status,
			cluster.StatusReason,
			pq.Array(cluster.AllowedTiers),
			pq.Array(cluster.Capabilities),
			cluster.Namespace,
			cluster.Region,
			cluster.EMRVirtualCluster,
			cluster.SparkServerURI).Scan(&id)

		if err != nil {
			span.SetTag("error", true)
			span.SetTag("error.msg", err.Error())
			return err
		}
		return nil
	} else {
		sql := `
			UPDATE cluster_state
			SET 
				name = $2,
				cluster_version = $3,
				status = $4,
				status_reason = $5,
				allowed_tiers = $6,
				capabilities = $7,
				namespace = $8,
				region = $9,
				emr_virtual_cluster = $10,
				spark_server_uri = $11,
				updated_at = NOW()
			WHERE id = $1;
		`
		result, err := sm.db.ExecContext(ctx, sql,
			cluster.ID,
			cluster.Name,
			cluster.ClusterVersion,
			cluster.Status,
			cluster.StatusReason,
			pq.Array(cluster.AllowedTiers),
			pq.Array(cluster.Capabilities),
			cluster.Namespace,
			cluster.Region,
			cluster.EMRVirtualCluster,
			cluster.SparkServerURI)

		if err != nil {
			span.SetTag("error", true)
			span.SetTag("error.msg", err.Error())
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			span.SetTag("error", true)
			span.SetTag("error.msg", err.Error())
			return err
		}

		if rows == 0 {
			span.SetTag("error", true)
			span.SetTag("error.msg", "Cluster not found")
			return exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Cluster with ID %s not found", cluster.ID),
			}
		}
		return nil
	}
}

func (sm *SQLStateManager) DeleteClusterMetadata(ctx context.Context, clusterID string) error {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.delete_cluster_metadata", "")
	defer span.Finish()
	span.SetTag("cluster.id", clusterID)
	sql := `DELETE FROM cluster_state WHERE id = $1`
	result, err := sm.db.ExecContext(ctx, sql, clusterID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return err
	}

	if count == 0 {
		span.SetTag("error", true)
		span.SetTag("error.msg", "Cluster not found")
		return exceptions.MissingResource{
			ErrorString: fmt.Sprintf("Cluster with ID %s not found", clusterID),
		}
	}
	return nil
}

func (sm *SQLStateManager) GetClusterByID(ctx context.Context, clusterID string) (ClusterMetadata, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_cluster_by_id", "")
	defer span.Finish()
	span.SetTag("cluster.id", clusterID)
	var cluster ClusterMetadata
	query := `
		SELECT 
			id, name, status, status_reason, status_since, allowed_tiers,
			capabilities, region, updated_at, namespace, emr_virtual_cluster, spark_server_uri
		FROM cluster_state 
		WHERE id = $1
	`
	err := sm.db.GetContext(ctx, &cluster, query, clusterID)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		if err == sql.ErrNoRows {
			return cluster, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Cluster with ID %s not found", clusterID),
			}
		}
		return cluster, err
	}

	// Add tags for the cluster data
	span.SetTag("cluster.name", cluster.Name)
	span.SetTag("cluster.status", cluster.Status)
	if cluster.ClusterVersion != "" {
		span.SetTag("cluster.version", cluster.ClusterVersion)
	}

	return cluster, nil
}

func ScanStringArray(arr *[]string, value interface{}) error {
	if value == nil {
		*arr = []string{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		var result []string
		if err := json.Unmarshal(v, &result); err == nil {
			*arr = result
			return nil
		}
		str := string(v)
		if len(str) < 2 {
			*arr = []string{}
			return nil
		}
		elements := strings.Split(str[1:len(str)-1], ",")
		result = make([]string, 0, len(elements))
		for _, e := range elements {
			if e != "" {
				// Remove quotes if they exist
				e = strings.Trim(e, "\"")
				result = append(result, e)
			}
		}
		*arr = result
		return nil
	default:
		return fmt.Errorf("unexpected type for string array: %T", value)
	}
}

func (arr *Tiers) Scan(value interface{}) error {
	if value == nil {
		*arr = Tiers{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		var result []string
		if err := json.Unmarshal(v, &result); err == nil {
			*arr = Tiers(result)
			return nil
		}
		str := string(v)
		if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
			*arr = Tiers{}
			return nil
		}
		str = str[1 : len(str)-1]
		if len(str) == 0 {
			*arr = Tiers{}
			return nil
		}
		elements := strings.Split(str, ",")
		result = make([]string, 0, len(elements))
		for _, e := range elements {
			if e == "" {
				continue
			}
			e = strings.Trim(e, "\"")
			result = append(result, e)
		}
		*arr = Tiers(result)
		return nil
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *Tiers", value)
	}
}

func (arr Tiers) Value() (driver.Value, error) {
	if len(arr) == 0 {
		return "{}", nil
	}
	quoted := make([]string, len(arr))
	for i, v := range arr {
		quoted[i] = fmt.Sprintf("\"%s\"", v)
	}
	return fmt.Sprintf("{%s}", strings.Join(quoted, ",")), nil
}

// Scan from db for Capabilities
func (arr *Capabilities) Scan(value interface{}) error {
	if value == nil {
		*arr = Capabilities{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		var result []string
		if err := json.Unmarshal(v, &result); err == nil {
			*arr = Capabilities(result)
			return nil
		}

		str := string(v)
		if len(str) < 2 {
			*arr = Capabilities{}
			return nil
		}
		elements := strings.Split(str[1:len(str)-1], ",")
		result = make([]string, 0, len(elements))
		for _, e := range elements {
			if e != "" {
				result = append(result, e)
			}
		}
		*arr = Capabilities(result)
		return nil
	default:
		return fmt.Errorf("unexpected type for string array: %T", value)
	}
}

// Value to db for Capabilities
func (arr Capabilities) Value() (driver.Value, error) {
	if len(arr) == 0 {
		return "{}", nil
	}
	return fmt.Sprintf("{%s}", strings.Join(arr, ",")), nil
}

func (sm *SQLStateManager) GetRunStatus(ctx context.Context, runID string) (RunStatus, error) {
	ctx, span := tracing.TraceJob(ctx, "flotilla.state.get_run_status", "")
	defer span.Finish()
	span.SetTag("job.run.id", runID)
	var status RunStatus

	tx, err := sm.db.BeginTx(ctx, nil)
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return status, errors.Wrap(err, "failed to begin transaction")
	}

	_, err = tx.ExecContext(ctx, "SET LOCAL lock_timeout = '500ms'")
	if err != nil {
		tx.Rollback()
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return status, errors.Wrap(err, "failed to set lock timeout")
	}

	err = tx.QueryRowContext(ctx, GetRunStatusSQL, runID).Scan(
		&status.RunID,
		&status.DefinitionID,
		&status.Alias,
		&status.ClusterName,
		&status.Status,
		&status.QueuedAt,
		&status.StartedAt,
		&status.FinishedAt,
		&status.ExitCode,
		&status.ExitReason,
		&status.Engine,
	)

	if err != nil {
		tx.Rollback()
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())

		if err == sql.ErrNoRows {
			return status, exceptions.MissingResource{
				ErrorString: fmt.Sprintf("Run with id %s not found", runID)}
		}

		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "55P03" {
			return status, exceptions.ConflictingResource{
				ErrorString: fmt.Sprintf("Run with id %s is currently locked, please retry", runID)}
		}

		return status, errors.Wrapf(err, "issue getting run status with id [%s]", runID)
	}

	err = tx.Commit()
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.msg", err.Error())
		return status, errors.Wrap(err, "failed to commit transaction")
	}

	//if status.Status != "" {
	//	span.SetTag("job.status", status.Status)
	//}

	return status, nil
}
