package state

//
// CreateTablesSQL postgres specific query for creating task
// definition, run, and related tables
//
const CreateTablesSQL = `
--
-- Definitions
--

CREATE TABLE IF NOT EXISTS task_def (
  definition_id character varying PRIMARY KEY,
  alias character varying,
  image character varying NOT NULL,
  group_name character varying NOT NULL,
  memory integer,
  cpu integer,
  gpu integer,
  command text,
  env jsonb,
  -- Refactor these
  "user" character varying,
  arn character varying,
  container_name character varying NOT NULL,
  task_type character varying,
  privileged boolean,
  -- Refactor these
  CONSTRAINT task_def_alias UNIQUE(alias)
);

CREATE TABLE IF NOT EXISTS task_def_ports (
  task_def_id character varying NOT NULL REFERENCES task_def(definition_id),
  port integer NOT NULL,
  CONSTRAINT task_def_ports_pkey PRIMARY KEY(task_def_id, port)
);

CREATE INDEX IF NOT EXISTS ix_task_def_alias ON task_def(alias);
CREATE INDEX IF NOT EXISTS ix_task_def_group_name ON task_def(group_name);
CREATE INDEX IF NOT EXISTS ix_task_def_image ON task_def(image);
CREATE INDEX IF NOT EXISTS ix_task_def_env ON task_def USING gin (env jsonb_path_ops);
--
-- Runs
--

CREATE TABLE IF NOT EXISTS task (
  run_id character varying NOT NULL PRIMARY KEY,
  definition_id character varying REFERENCES task_def(definition_id),
  alias character varying,
  image character varying,
  cluster_name character varying,
  exit_code integer,
  exit_reason character varying,
  status character varying,
  queued_at timestamp with time zone,
  started_at timestamp with time zone,
  finished_at timestamp with time zone,
  instance_id character varying,
  instance_dns_name character varying,
  group_name character varying,
  env jsonb,
  -- Refactor these --
  task_arn character varying,
  docker_id character varying,
  "user" character varying,
  task_type character varying,
  -- Refactor these --
  command text,
  memory integer,
  cpu integer,
  gpu integer,
  ephemeral_storage integer,
  node_lifecycle text,
  engine character varying DEFAULT 'ecs' NOT NULL,
  container_name text,
  pod_name text,
  namespace text,
  max_cpu_used integer,
  max_memory_used integer,
  pod_events jsonb
);

CREATE INDEX IF NOT EXISTS ix_task_definition_id ON task(definition_id);
CREATE INDEX IF NOT EXISTS ix_task_cluster_name ON task(cluster_name);
CREATE INDEX IF NOT EXISTS ix_task_status ON task(status);
CREATE INDEX IF NOT EXISTS ix_task_group_name ON task(group_name);
CREATE INDEX IF NOT EXISTS ix_task_env ON task USING gin (env jsonb_path_ops);
CREATE INDEX IF NOT EXISTS ix_task_definition_id ON task(definition_id);
CREATE INDEX IF NOT EXISTS ix_task_task_arn ON task(task_arn);
CREATE INDEX IF NOT EXISTS ix_task_definition_id_started_at_desc ON task(definition_id, started_at DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS ix_task_definition_id_started_at_desc_engine ON task(definition_id, started_at DESC NULLS LAST, engine);
--
-- Status
--

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

--
-- Tags
--
CREATE TABLE IF NOT EXISTS tags (
  text character varying NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS task_def_tags (
  tag_id character varying NOT NULL REFERENCES tags(text),
  task_def_id character varying NOT NULL REFERENCES task_def(definition_id)
);

CREATE TABLE IF NOT EXISTS worker (
  worker_type character varying,
  engine character varying,
  count_per_instance integer
);
`

//
// DefinitionSelect postgres specific query for definitions
//
const DefinitionSelect = `
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
  ports                     as ports,
  tags                      as tags,
  td.privileged             as privileged,
  td.cpu                    as cpu,
  td.gpu                    as gpu
  from (select * from task_def) td left outer join
    (select task_def_id,
      array_to_json(array_agg(port))::TEXT as ports
        from task_def_ports group by task_def_id
    ) tdp
  on td.definition_id = tdp.task_def_id left outer join
    (select task_def_id,
      array_to_json(array_agg(tag_id))::TEXT as tags
        from task_def_tags group by task_def_id
    ) tdt
  on td.definition_id = tdt.task_def_id
`

//
// ListDefinitionsSQL postgres specific query for listing definitions
//
const ListDefinitionsSQL = DefinitionSelect + "\n%s %s limit $1 offset $2"

//
// GetDefinitionSQL postgres specific query for getting a single definition
//
const GetDefinitionSQL = DefinitionSelect + "\nwhere definition_id = $1"

//
// GetDefinitionByAliasSQL get definition by alias
//
const GetDefinitionByAliasSQL = DefinitionSelect + "\nwhere alias = $1"

//
// RunSelect postgres specific query for runs
//
const RunSelect = `
select
  coalesce(t.task_arn,'')                    as taskarn,
  t.run_id                                   as runid,
  coalesce(t.definition_id,'')               as definitionid,
  coalesce(t.alias,'')                       as alias,
  coalesce(t.image,'')                       as image,
  coalesce(t.cluster_name,'')                as clustername,
  t.exit_code                                as exitcode,
  t.exit_reason                              as exitreason,
  coalesce(t.status,'')                      as status,
  queued_at                                  as queuedat,
  started_at                                 as startedat,
  finished_at                                as finishedat,
  coalesce(t.instance_id,'')                 as instanceid,
  coalesce(t.instance_dns_name,'')           as instancednsname,
  coalesce(t.group_name,'')                  as groupname,
  coalesce(t.user,'')                        as "user",
  coalesce(t.task_type,'')                   as tasktype,
  env::TEXT                                  as env,
  command,
  memory,
  cpu,
  gpu,
  engine,
  ephemeral_storage as ephemeralstorage,
  node_lifecycle as nodelifecycle,
  container_name as containername,
  pod_name as podname,
  namespace,
  max_cpu_used as maxcpuused,
  max_memory_used as maxmemoryused,
  pod_events::TEXT as podevents
from task t
`

//
// ListRunsSQL postgres specific query for listing runs
//
const ListRunsSQL = RunSelect + "\n%s %s limit $1 offset $2"

//
// GetRunSQL postgres specific query for getting a single run
//
const GetRunSQL = RunSelect + "\nwhere run_id = $1"

//
// GetRunSQLForUpdate postgres specific query for getting a single run
// for update
//
const GetRunSQLForUpdate = GetRunSQL + " for update"

//
// GroupsSelect postgres specific query for getting existing definition
// group_names
//
const GroupsSelect = `
select distinct group_name from task_def
`

//
// TagsSelect postgres specific query for getting existing definition tags
//
const TagsSelect = `
select distinct text from tags
`

//
// ListGroupsSQL postgres specific query for listing definition group_names
//
const ListGroupsSQL = GroupsSelect + "\n%s order by group_name asc limit $1 offset $2"

//
// ListTagsSQL postgres specific query for listing definition tags
//
const ListTagsSQL = TagsSelect + "\n%s order by text asc limit $1 offset $2"

//
// WorkerSelect postgres specific query for workers
//
const WorkerSelect = `
  select
    worker_type        as workertype,
    count_per_instance as countperinstance,
    engine
  from worker
`

//
// ListWorkersSQL postgres specific query for listing workers
//
const ListWorkersSQL = WorkerSelect

const GetWorkerEngine = WorkerSelect + "\nwhere engine = $1"


//
// GetWorkerSQL postgres specific query for retrieving data for a specific
// worker type.
//
const GetWorkerSQL = WorkerSelect + "\nwhere worker_type = $1 and engine = $2"

//
// GetWorkerSQLForUpdate postgres specific query for retrieving data for a specific
// worker type; locks the row.
//
const GetWorkerSQLForUpdate = GetWorkerSQL + " for update"
