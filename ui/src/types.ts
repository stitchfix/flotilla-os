import { Omit } from "lodash"

export type Env = {
  name: string
  value: any
}

export type Task = {
  env: Env[]
  arn: string
  definition_id: string
  group_name: string
  alias: string
  command: string
  shared_memory_size?: number
} & ExecutableResources

export type RunInstance = {
  dns_name: string
  instance_id: string
}

export type CloudTrailNotifications = {}

export enum ExecutableType {
  ExecutableTypeDefinition = "task_definition",
  ExecutableTypeTemplate = "template",
}

export type Run = {
  attempt_count: number
  alias: string
  cluster: string
  command?: string
  cpu: number
  cpu_limit: number
  definition_id: string
  env: Env[]
  exit_code?: number
  exit_reason?: string
  finished_at?: string
  gpu?: number
  group_name: string
  image: string
  instance: RunInstance
  memory: number
  memory_limit: number
  queued_at: string | undefined
  run_id: string
  started_at?: string
  status: RunStatus
  task_arn: string
  engine: ExecutionEngine
  node_lifecycle?: NodeLifecycle
  max_cpu_used: number | null | undefined
  max_memory_used: number | null | undefined
  pod_name: string | null | undefined
  cloudtrail_notifications: CloudtrailRecords
  executable_id: string
  executable_type: ExecutableType
  execution_request_custom: any
}

export type RunLog = {
  log: string
  last_seen?: string
}

export type RunLogRaw = string

//
// Enums
//

export enum HTTPMethod {
  GET = "get",
  PUT = "put",
  POST = "post",
  DELETE = "delete",
}

export enum SortOrder {
  ASC = "asc",
  DESC = "desc",
}

export enum RunStatus {
  PENDING = "PENDING",
  QUEUED = "QUEUED",
  RUNNING = "RUNNING",
  STOPPED = "STOPPED",
  NEEDS_RETRY = "NEEDS_RETRY",
}

export enum EnhancedRunStatus {
  PENDING = "PENDING",
  QUEUED = "QUEUED",
  RUNNING = "RUNNING",
  STOPPED = "STOPPED",
  NEEDS_RETRY = "NEEDS_RETRY",
  SUCCESS = "SUCCESS",
  FAILED = "FAILED",
}

// 3rd party

export type SelectOption = { label: string; value: string }

export type SelectProps = {
  value: string
  onChange: (value: string) => void
  isDisabled: boolean
}

export type MultiSelectProps = {
  value: string[]
  onChange: (value: string[]) => void
  isDisabled: boolean
}

//
// Request/Response
// These type definitions relate to the arguments required to invoke methods
// of the Flotilla client and the response the server returns.
//
export type RequestArgs = {
  method: HTTPMethod
  url: string
  params?: object
  data?: any
}

export type ListRequestArgs = {
  offset: number
  limit: number
  sort_by?: string
  order?: SortOrder
}

export type ListResponse = {
  offset: number
  limit: number
  sort_by?: string
  order?: SortOrder
  total: number
}

export type ListTaskRunsParams = Omit<ListRunParams, "alias">
export type ListTaskRunsResponse = Omit<ListRunResponse, "alias">

export type ListTaskParams = ListRequestArgs & {
  alias?: string[]
  group_name?: string[]
  image?: string[]
}

export type ListTaskResponse = ListResponse & {
  definitions: Task[]
  alias?: string[]
  group_name?: string[]
  image?: string[]
}

export type CreateTaskPayload = UpdateTaskPayload & { alias: string }

export type UpdateTaskPayload = {
  env: Env[]
  image: string
  group_name: string
  memory: number
  cpu: number
  command: string
  tags: string[]
  adaptive_resource_allocation?: boolean
}

export enum ExecutionEngine {
  ECS = "ecs",
  EKS = "eks",
}

export const DefaultExecutionEngine = ExecutionEngine.EKS

export enum NodeLifecycle {
  SPOT = "spot",
  ON_DEMAND = "ondemand",
}

export const DefaultNodeLifecycle = NodeLifecycle.SPOT

export type LaunchRequestV2 = {
  cluster: string
  command?: string | null
  cpu?: number
  engine: ExecutionEngine
  env?: Env[]
  memory?: number
  node_lifecycle?: NodeLifecycle
  owner_id?: string
  run_tags?: { [key: string]: any }
}

export type ListRunParams = ListRequestArgs & {
  env?: string[]
  cluster_name?: string
  alias?: string[]
  status?: RunStatus
}

export type ListRunResponse = ListResponse & {
  history: Run[]
  env_filters?: { [name: string]: any }
  cluster_name?: string
  alias?: string[]
  status?: RunStatus
}

export type ListClustersResponse = ListResponse & { clusters: string[] | null }
export type ListGroupsResponse = ListResponse & { groups: string[] | null }
export type ListTagsResponse = ListResponse & { tags: string[] | null }

export type FieldSpec = {
  name: string
  label: string
  description: string
  initialValue: any
}

export type PodEvent = {
  timestamp: string
  event_type: string
  reason: string
  source_object: string
  message: string
}

export type ListRunEventsResponse = {
  total: number
  pod_events: PodEvent[] | null
}

export enum RunTabId {
  LOGS = "l",
  EVENTS = "e",
  CLOUDTRAIL = "ct",
  METADATA = "md",
}

export type ExecutableResources = {
  image: string
  memory: number
  gpu: number
  cpu: number
  env: Env[]
  privileged: boolean
  adaptive_resource_allocation: boolean
  container_name: string
  tags: string[]
}

export type Template = {
  template_id: string
  template_name: string
  version: number
  schema: object
  command_template: string
  defaults: object | null | undefined
  avatar_uri: string | null | undefined
} & ExecutableResources

export type ListTemplateParams = ListRequestArgs & {
  latest_only?: boolean
}

export type ListTemplateResponse = ListResponse & {
  templates: Template[]
}

export type TemplateExecutionRequest = {
  template_payload: object
} & ExecutionRequestCommon

export type ExecutionRequestCommon = {
  cluster: string
  command?: string
  cpu?: number
  engine: ExecutionEngine
  env?: Env[]
  memory?: number
  node_lifecycle?: NodeLifecycle
  owner_id: string
}
export type CloudtrailRecord = {
  eventSource: string
  eventName: string
}

export type CloudtrailRecords = {
  Records: CloudtrailRecord[] | null
}

export const EnhancedRunStatusEmojiMap: Map<
  EnhancedRunStatus,
  string
> = new Map([
  [EnhancedRunStatus.PENDING, "🕒"],
  [EnhancedRunStatus.QUEUED, "🕒"],
  [EnhancedRunStatus.RUNNING, "🌀"],
  [EnhancedRunStatus.SUCCESS, "✅"],
  [EnhancedRunStatus.FAILED, "❌"],
  [EnhancedRunStatus.NEEDS_RETRY, "❌"],
])

export type ListTemplateHistoryParams = Omit<ListRunParams, "alias">
export type ListTemplateHistoryResponse = Omit<ListRunResponse, "alias">
export type Executable = Template | Task
