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
  adaptive_resource_allocation boolean,
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
  command_hash text,
  memory integer,
  cpu integer,
  gpu integer,
  ephemeral_storage integer,
  node_lifecycle text,
  engine character varying DEFAULT 'eks' NOT NULL,
  container_name text,
  pod_name text,
  namespace text,
  max_cpu_used integer,
  max_memory_used integer,
  pod_events jsonb,
  cloudtrail_notifications jsonb
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
