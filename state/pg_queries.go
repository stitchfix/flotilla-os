package state

//
// DefinitionSelect postgres specific query for definitions
//
const DefinitionSelect = `
select
  coalesce(td.arn,'')       as arn,
  td.definition_id          as definitionid,
  td.adaptive_resource_allocation as adaptiveresourceallocation,
  td.image                  as image,
  td.group_name             as groupname,
  td.container_name         as containername,
  coalesce(td.user,'')      as "user",
  td.alias                  as alias,
  td.memory                 as memory,
  coalesce(td.command,'')   as command,
  env::TEXT                 as env,
  ports                     as ports,
  tags                      as tags,
  td.privileged             as privileged,
  td.cpu                    as cpu,
  td.gpu                    as gpu,
  coalesce(td.template_id,'') as templateid,
  td.template_payload::TEXT          as templatepayload
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

const TaskResourcesSelectCommandSQL = `
SELECT cast((percentile_disc(0.99) within GROUP (ORDER BY A.max_memory_used)) * 1.5 as int) as memory,
       cast((percentile_disc(0.99) within GROUP (ORDER BY A.max_cpu_used)) * 1.25  as int)  as cpu
FROM (SELECT max_memory_used, max_cpu_used
      FROM TASK
      WHERE definition_id = $1
           AND exit_code = 0
           AND engine = 'eks'
           AND max_memory_used is not null
           AND max_cpu_used is not null
           AND command_hash is not NULL
           AND queued_at >= CURRENT_TIMESTAMP - INTERVAL '7 days'
           AND command_hash = (SELECT command_hash FROM task WHERE run_id = $2)
      LIMIT 30) A
`

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
  pod_events::TEXT as podevents,
  command_hash as commandhash,
  coalesce(t.template_id,'') as templateid,
  t.template_payload::TEXT          as templatepayload
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

// definitionTemplateSelect retrives definition template data.
const definitionTemplateSelect = `
  SELECT id as templateid, type, version, schema, template, image
  FROM definition_template
`
const ListDefinitionTemplateLatestOnlySQL = `
  SELECT DISTINCT ON (type)
    id as templateid, type, version, schema, template, image
  FROM definition_template
  ORDER BY type, version DESC, id
  LIMIT $1 OFFSET $2;
`
const ListDefinitionTemplateSQL = definitionTemplateSelect + "\n limit $1 offset $2"
const GetDefinitionTemplateByIdSQL = definitionTemplateSelect + "\n where id = $1"
