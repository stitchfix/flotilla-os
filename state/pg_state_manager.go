package state

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"strings"
    "encoding/json"
    "database/sql/driver"
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
    select
      coalesce(td.arn,'')       as arn,
      td.definition_id          as definitionid,
      td.image                  as image,
      td.group_name             as groupname,
      td.container_name         as containername,
      coalesce(td.user,'')      as "user",
      td.alias                  as alias,
      td.memory                 as memory,
      coalesce(td.command,'')   as command,
      coalesce(td.task_type,'') as tasktype,
      env::TEXT                 as env,
      ports                     as ports
    from (select * from task_def) td left outer join
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

func (sm *SQLStateManager) GetDefinition(definitionID string) (Definition, error) {
    var err error
    var definition Definition

    sql := `
    select
      coalesce(td.arn,'')       as arn,
      td.definition_id          as definitionid,
      td.image                  as image,
      td.group_name             as groupname,
      td.container_name         as containername,
      coalesce(td.user,'')      as "user",
      td.alias                  as alias,
      td.memory                 as memory,
      coalesce(td.command,'')   as command,
      coalesce(td.task_type,'') as tasktype,
      env::TEXT                 as env,
      ports                     as ports
    from (select * from task_def) td left outer join
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
      where definition_id = $1
    `
    err = sm.db.Get(&definition, sql, definitionID)
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
        for _, e := range (*existing.Env) {
            if _, err = tx.Exec(insertEnv, definitionID, e.Name, e.Value); err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    if existing.Ports != nil {
        for _, p := range (*existing.Ports) {
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
        for _, e := range (*d.Env) {
            if _, err = tx.Exec(insertEnv, d.DefinitionID, e.Name, e.Value); err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    if d.Ports != nil {
        for _, p := range (*d.Ports) {
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

    for _, stmt := range(statements) {
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

    template := `
    select
      coalesce(t.task_arn,'')                    as taskarn,
      t.run_id                                   as runid,
      coalesce(t.definition_id,'')               as definitionid,
      coalesce(t.cluster_name,'')                as clustername,
      t.exit_code                                as exitcode,
      coalesce(t.status,'')                      as status,
      coalesce(t.started_at, DATE '0001-01-01')  as startedat,
      coalesce(t.finished_at, DATE '0001-01-01') as finishedat,
      coalesce(t.instance_id,'')                 as instanceid,
      coalesce(t.instance_dns_name,'')           as instancednsname,
      coalesce(t.group_name,'')                  as groupname,
      coalesce(t.user,'')                        as "user",
      coalesce(t.task_type,'')                   as tasktype,
      env::TEXT                                  as env
    from (select * from task) t left outer join
      (select task_id,
        array_to_json(
          array_agg(json_build_object('name',name,'value',coalesce(value,'')))) as env
            from task_environments group by task_id
      ) te
    on t.run_id = te.task_id
      %s %s limit $1 offset $2
    `

    sql := fmt.Sprintf(template, whereClause, orderQuery)
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

func (r *Run) ValidOrderField(field string) bool {
    for _, f := range r.ValidOrderFields() {
        if field == f {
            return true
        }
    }
    return false
}

func (r *Run) ValidOrderFields() []string {
    return []string{"run_id","cluster_name","status","started_at","finished_at","group_name"}
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