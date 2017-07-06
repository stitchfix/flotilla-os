package state

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
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

//
// Definitions
//

func (sm *SQLStateManager) ListDefinitions() {

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
