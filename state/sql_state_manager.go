package state

import (
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    _ "github.com/mattn/go-sqlite3"
)

type SQLStateManager struct {
    db *sqlx.DB
}

func (sm *SQLStateManager) Initialize() error {
    var err error
    if sm.db, err = sqlx.Connect("sqlite3", ":memory:"); err != nil {
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