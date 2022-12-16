package state

// DefinitionSelect postgres specific query for definitions
const DefinitionSelect = `
select td.definition_id                    as definitionid,
       td.adaptive_resource_allocation     as adaptiveresourceallocation,
       td.image                            as image,
       td.group_name                       as groupname,
       td.alias                            as alias,
       td.memory                           as memory,
       coalesce(td.command, '')            as command,
       coalesce(td.task_type, '')          as tasktype,
       env::TEXT                           as env,
       td.cpu                              as cpu,
       td.gpu                              as gpu,
       array_to_json('{""}'::TEXT[])::TEXT as tags,
       array_to_json('{}'::INT[])::TEXT    as ports
from (select * from task_def) td
`

// ListDefinitionsSQL postgres specific query for listing definitions
const ListDefinitionsSQL = DefinitionSelect + "\n%s %s limit $1 offset $2"

// GetDefinitionSQL postgres specific query for getting a single definition
const GetDefinitionSQL = DefinitionSelect + "\nwhere definition_id = $1"

// GetDefinitionByAliasSQL get definition by alias
const GetDefinitionByAliasSQL = DefinitionSelect + "\nwhere alias = $1"

const TaskResourcesSelectCommandSQL = `
SELECT cast((percentile_disc(0.99) within GROUP (ORDER BY A.max_memory_used)) * 1.75 as int) as memory,
       cast((percentile_disc(0.99) within GROUP (ORDER BY A.max_cpu_used)) * 1.25  as int)  as cpu
FROM (SELECT memory as max_memory_used, cpu as max_cpu_used
      FROM TASK
      WHERE
           queued_at >= CURRENT_TIMESTAMP - INTERVAL '3 days'
           AND (exit_code = 137 or exit_reason = 'OOMKilled')
           AND engine = 'eks'
           AND definition_id = $1
           AND command_hash = (SELECT command_hash FROM task WHERE run_id = $2)
      LIMIT 30) A
`

const TaskResourcesExecutorCountSQL = `
SELECT least(coalesce(cast((percentile_disc(0.99) within GROUP (ORDER BY A.executor_count)) as int), 25), 100) as executor_count
FROM (SELECT CASE
                 WHEN (exit_reason like '%Exception%')
                     THEN (spark_extension -> 'spark_submit_job_driver' -> 'num_executors')::int * 1.75
                 ELSE (spark_extension -> 'spark_submit_job_driver' -> 'num_executors')::int * 1
                 END as executor_count
      FROM TASK
      WHERE
           queued_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
           AND engine = 'eks-spark'
           AND definition_id = $1
           AND command_hash = $2
      LIMIT 30) A
`
const TaskResourcesDriverOOMSQL = `
SELECT (spark_extension -> 'driver_oom')::boolean AS driver_oom
FROM TASK
WHERE queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
  AND engine = 'eks-spark'
  AND definition_id = $1
  AND command_hash = $2
  AND exit_code = 137
  AND spark_extension ? 'driver_oom'
GROUP BY 1
`

const TaskIdempotenceKeyCheckSQL = `
WITH runs as (
    SELECT run_id
    FROM task
    WHERE idempotence_key = $1
      and (exit_code = 0 or exit_code is null)
      and queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days')
SELECT run_id
FROM runs
LIMIT 1;
`

const TaskResourcesExecutorOOMSQL = `
SELECT CASE WHEN A.c >= 1 THEN true::boolean ELSE false::boolean END
FROM (SELECT count(*) as c
      FROM TASK
      WHERE
           queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
           AND definition_id = $1
           AND command_hash = $2
		   AND engine = 'eks-spark'
           AND exit_code !=0
      LIMIT 30) A
`

const TaskResourcesExecutorNodeLifecycleSQL = `
SELECT CASE WHEN A.c >= 1 THEN 'ondemand' ELSE 'spot' END
FROM (SELECT count(*) as c
      FROM TASK
      WHERE
           queued_at >= CURRENT_TIMESTAMP - INTERVAL '12 hour'
           AND definition_id = $1
           AND command_hash = $2
           AND exit_code !=0
      LIMIT 30) A
`

const TaskExecutionRuntimeCommandSQL = `
SELECT percentile_disc(0.95) within GROUP (ORDER BY A.minutes) as minutes
FROM (SELECT EXTRACT(epoch from finished_at - started_at) / 60 as minutes
      FROM TASK
      WHERE definition_id = $1
        AND exit_code = 0
        AND engine = 'eks'
        AND queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
        AND command_hash = (SELECT command_hash FROM task WHERE run_id = $2)
      LIMIT 30) A
`

const ListFailingNodesSQL = `
SELECT instance_dns_name
FROM (
         SELECT instance_dns_name, count(*) as c
         FROM TASK
         WHERE (exit_code = 128 OR
                pod_events @> '[{"reason": "Failed"}]' OR
                pod_events @> '[{"reason": "FailedSync"}]' OR
                pod_events @> '[{"reason": "FailedCreatePodSandBox"}]' OR
                pod_events @> '[{"reason": "OutOfmemory"}]')
           AND engine = 'eks'
           AND queued_at >= NOW() - INTERVAL '1 HOURS'
           AND instance_dns_name like 'ip-%'
         GROUP BY 1
         order by 2 desc) AS all_nodes
WHERE c >= 5
`

const PodReAttemptRate = `
SELECT (multiple_attempts / (CASE WHEN single_attempts = 0 THEN 1 ELSE single_attempts END)) AS attempts
FROM (
      SELECT COUNT(CASE WHEN attempt_count <= 1 THEN 1 END) * 1.0 AS single_attempts,
             COUNT(CASE WHEN attempt_count > 1 THEN 1 END) * 1.0 AS multiple_attempts
      FROM task
      WHERE engine = 'eks' AND
            queued_at >= NOW() - INTERVAL '18 MINUTES' AND
            node_lifecycle = 'spot') A
`

// RunSelect postgres specific query for runs
const RunSelect = `
select t.run_id                          as runid,
       coalesce(t.definition_id, '')     as definitionid,
       coalesce(t.alias, '')             as alias,
       coalesce(t.image, '')             as image,
       coalesce(t.cluster_name, '')      as clustername,
       t.exit_code                       as exitcode,
       t.exit_reason                     as exitreason,
       coalesce(t.status, '')            as status,
       queued_at                         as queuedat,
       started_at                        as startedat,
       finished_at                       as finishedat,
       coalesce(t.instance_id, '')       as instanceid,
       coalesce(t.instance_dns_name, '') as instancednsname,
       coalesce(t.group_name, '')        as groupname,
       coalesce(t.task_type, '')         as tasktype,
       env::TEXT                         as env,
       command,
       memory,
       cpu,
       gpu,
       engine,
       ephemeral_storage                 as ephemeralstorage,
       node_lifecycle                    as nodelifecycle,
       pod_name                          as podname,
       namespace,
       max_cpu_used                      as maxcpuused,
       max_memory_used                   as maxmemoryused,
       pod_events::TEXT                  as podevents,
       command_hash                      as commandhash,
       cloudtrail_notifications::TEXT    as cloudtrailnotifications,
       coalesce(executable_id, '')       as executableid,
       coalesce(executable_type, '')     as executabletype,
       execution_request_custom::TEXT    as executionrequestcustom,
       cpu_limit                         as cpulimit,
       memory_limit                      as memorylimit,
       attempt_count                     as attemptcount,
       spawned_runs::TEXT                as spawnedruns,
       run_exceptions::TEXT              as runexceptions,
       active_deadline_seconds           as activedeadlineseconds,
       spark_extension::TEXT             as sparkextension,
       metrics_uri                       as metricsuri,
       description                       as description,
	   idempotence_key                   as idempotencekey,
       coalesce("user", '')              as user,
	   coalesce(arch, '')                as arch,
	   labels::TEXT                      as labels
from task t
`

// ListRunsSQL postgres specific query for listing runs
const ListRunsSQL = RunSelect + "\n%s %s limit $1 offset $2"

// GetRunSQL postgres specific query for getting a single run
const GetRunSQL = RunSelect + "\nwhere run_id = $1"

const GetRunSQLByEMRJobId = RunSelect + "\nwhere spark_extension->>'emr_job_id' = $1"

// GetRunSQLForUpdate postgres specific query for getting a single run
// for update
const GetRunSQLForUpdate = GetRunSQL + " for update"

// GroupsSelect postgres specific query for getting existing definition
// group_names
const GroupsSelect = `
select distinct group_name from task_def
`

// TagsSelect postgres specific query for getting existing definition tags
const TagsSelect = `
select distinct text from tags
`

// ListGroupsSQL postgres specific query for listing definition group_names
const ListGroupsSQL = GroupsSelect + "\n%s order by group_name asc limit $1 offset $2"

// ListTagsSQL postgres specific query for listing definition tags
const ListTagsSQL = TagsSelect + "\n%s order by text asc limit $1 offset $2"

// WorkerSelect postgres specific query for workers
const WorkerSelect = `
  select
    worker_type        as workertype,
    count_per_instance as countperinstance,
    engine
  from worker
`

// ListWorkersSQL postgres specific query for listing workers
const ListWorkersSQL = WorkerSelect

const GetWorkerEngine = WorkerSelect + "\nwhere engine = $1"

// GetWorkerSQL postgres specific query for retrieving data for a specific
// worker type.
const GetWorkerSQL = WorkerSelect + "\nwhere worker_type = $1 and engine = $2"

// GetWorkerSQLForUpdate postgres specific query for retrieving data for a specific
// worker type; locks the row.
const GetWorkerSQLForUpdate = GetWorkerSQL + " for update"

// TemplateSelect selects a template
const TemplateSelect = `
SELECT
  template_id as templateid,
  template_name as templatename,
  version,
  schema,
  command_template as commandtemplate,
  adaptive_resource_allocation as adaptiveresourceallocation,
  image,
  memory,
  env::TEXT as env,
  privileged,
  cpu,
  gpu,
  defaults,
  coalesce(avatar_uri, '') as avataruri
FROM template
`

// ListTemplatesSQL postgres specific query for listing templates
const ListTemplatesSQL = TemplateSelect + "\n%s limit $1 offset $2"

// GetTemplateByIDSQL postgres specific query for getting a single template
const GetTemplateByIDSQL = TemplateSelect + "\nwhere template_id = $1"

// ListTemplatesLatestOnlySQL lists the latest version of each distinct
// template name.
const ListTemplatesLatestOnlySQL = `
  SELECT DISTINCT ON (template_name)
    template_id as templateid,
    template_name as templatename,
    version,
    schema,
    command_template as commandtemplate,
    adaptive_resource_allocation as adaptiveresourceallocation,
    image,
    memory,
    env::TEXT as env,
    privileged,
    cpu,
    gpu,
    defaults,
    coalesce(avatar_uri, '') as avataruri
  FROM template
  ORDER BY template_name, version DESC, template_id
  LIMIT $1 OFFSET $2
`

// GetTemplateLatestOnlySQL get the latest version of a specific template name.
const GetTemplateLatestOnlySQL = TemplateSelect + "\nWHERE template_name = $1 ORDER BY version DESC LIMIT 1;"
const GetTemplateByVersionSQL = TemplateSelect + "\nWHERE template_name = $1 AND version = $2 ORDER BY version DESC LIMIT 1;"
