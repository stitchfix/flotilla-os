package state

import (
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
	var err error
	if err = sm.createDefinitionTable(); err != nil {
		return err
	}
	if err = sm.createDefinitionEnvTable(); err != nil {
		return err
	}
	if err = sm.createDefinitionPortsTable(); err != nil {
		return err
	}
	if err = sm.createTaskTable(); err != nil {
		return err
	}
	if err = sm.createTaskEnvTable(); err != nil {
		return err
	}
	if err = sm.createTaskStatusTable(); err != nil {
		return err
	}
	return nil
}

func (sm *SQLStateManager) createDefinitionTable() error {
	ddl := `
    CREATE TABLE IF NOT EXISTS task_def (
      definition_id character varying PRIMARY KEY,
      alias character varying,
      image character varying NOT NULL,
      group_name character varying NOT NULL,
      memory integer,
      command text,
      -- Refactor these
      "user" character varying,
      arn character varying,
      container_name character varying NOT NULL,
      task_type character varying,
      -- Refactor these
      CONSTRAINT task_def_alias UNIQUE(alias)
    );

    CREATE INDEX IF NOT EXISTS ix_task_def_alias ON task_def(alias);
    CREATE INDEX IF NOT EXISTS ix_task_def_group_name ON task_def(group_name);
    CREATE INDEX IF NOT EXISTS ix_task_def_image ON task_def(image);
    `
	_, err := sm.db.Exec(ddl)
	return err
}

func (sm *SQLStateManager) createDefinitionEnvTable() error {
	ddl := `
    CREATE TABLE IF NOT EXISTS task_def_environments (
      task_def_id character varying NOT NULL REFERENCES task_def(definition_id),
      name character varying NOT NULL,
      value character varying,
      CONSTRAINT task_def_environments_pkey PRIMARY KEY(task_def_id, name)
    );
    `
	_, err := sm.db.Exec(ddl)
	return err
}

func (sm *SQLStateManager) createDefinitionPortsTable() error {
	ddl := `
	CREATE TABLE IF NOT EXISTS task_def_ports (
      task_def_id character varying NOT NULL REFERENCES task_def(definition_id),
      port integer NOT NULL,
      CONSTRAINT task_def_ports_pkey PRIMARY KEY(task_def_id, port)
    );
	`
	_, err := sm.db.Exec(ddl)
	return err
}

func (sm *SQLStateManager) createTaskTable() error {
	ddl := `
	CREATE TABLE IF NOT EXISTS task (
	  run_id character varying NOT NULL PRIMARY KEY,
	  definition_id character varying REFERENCES task_def(definition_id),
	  cluster_name character varying,
	  exit_code integer,
	  status character varying,
	  started_at timestamp with time zone,
	  finished_at timestamp with time zone,
	  instance_id character varying,
      instance_dns_name character varying,
	  group_name character varying,
	  -- Refactor these --
      task_arn character varying,
      docker_id character varying,
      "user" character varying,
      task_type character varying
      -- Refactor these --
    );

    CREATE INDEX IF NOT EXISTS ix_task_cluster_name ON task(cluster_name);
    CREATE INDEX IF NOT EXISTS ix_task_status ON task(status);
    CREATE INDEX IF NOT EXISTS ix_task_group_name ON task(group_name);
	`
	_, err := sm.db.Exec(ddl)
	return err
}

func (sm *SQLStateManager) createTaskEnvTable() error {
	ddl := `
    CREATE TABLE IF NOT EXISTS task_environments (
      task_id character varying NOT NULL REFERENCES task(run_id),
      name character varying NOT NULL,
      value character varying,
      CONSTRAINT task_environments_pkey PRIMARY KEY(task_id, name)
    );
    `
	_, err := sm.db.Exec(ddl)
	return err
}

func (sm *SQLStateManager) createTaskStatusTable() error {

	ddl := `
	CREATE TABLE IF NOT EXISTS task_status (
      status_id integer NOT NULL PRIMARY KEY,
      task_arn character varying,
      status_version integer NOT NULL,
      status character varying,
      "timestamp" timestamp with time zone DEFAULT now()
    );

	CREATE INDEX IF NOT EXISTS ix_task_status_task_arn ON task_status(task_arn);

    CREATE SEQUENCE IF NOT EXISTS task_status_status_id_seq
      START WITH 1
      INCREMENT BY 1
      NO MINVALUE
      NO MAXVALUE
      CACHE 1;

    ALTER TABLE ONLY task_status ALTER COLUMN status_id SET DEFAULT nextval('task_status_status_id_seq'::regclass);
	`
	_, err := sm.db.Exec(ddl)
	return err
}

func (sm *SQLStateManager) makeWhereClause(filters map[string]string) []string {
	wc := make([]string, len(filters))
	i := 0
	for k, v := range filters {
		fmtString := "%s=%s"
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

	template := `
    select td.*, env::TEXT as env, ports as ports from (select * from task_def) td left outer join
      (select task_def_id,
        array_to_json(
          array_agg(json_build_object('name',name,'value',coalesce(value,'')))) as env
            from task_def_environments group by task_def_id
      ) tde
    on td.definition_id = tde.task_def_id left outer join
      (select task_def_id,
        array_to_json(array_agg(port))::TEXT as ports
          from task_def_ports group by task_def_id
      ) tdp
    on td.definition_id = tdp.task_def_id
      %s %s limit $1 offset $2
    `

	sql := fmt.Sprintf(template, whereClause, orderQuery)
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

func (sm *SQLStateManager) GetDefinition() {
}

func (sm *SQLStateManager) UpdateDefinition() {

}

func (sm *SQLStateManager) CreateDefinition() {

}

func (sm *SQLStateManager) DeleteDefinition() {

}

//
// Runs
//

func (sm *SQLStateManager) ListRuns() {

}

func (sm *SQLStateManager) GetRun() {
}

func (sm *SQLStateManager) UpdateRun() {

}

func (sm *SQLStateManager) CreateRun() {

}

func (sm *SQLStateManager) Cleanup() {

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
