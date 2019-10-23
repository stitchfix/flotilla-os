import { Omit } from "lodash"

export type Env = {
  name: string
  value: any
}

export type Task = {
  env: Env[]
  arn: string
  definition_id: string
  image: string
  group_name: string
  container_name: string
  alias: string
  memory: number
  cpu: number
  command: string
  tags: string[]
  privileged: boolean
}

export type RunInstance = {
  dns_name: string
  instance_id: string
}

export type Run = {
  instance: RunInstance
  task_arn: string
  run_id: string
  definition_id: string
  alias: string
  image: string
  cluster: string
  exit_code?: number
  exit_reason?: string
  status: RunStatus
  started_at?: string
  finished_at?: string
  group_name: string
  env: Env[]
}

export type RunLog = {
  log: string
  last_seen?: string
}

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
}

export type MultiSelectProps = {
  value: string[]
  onChange: (value: string[]) => void
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
}

export type RunTaskPayload = {
  cluster: string
  env?: Env[]
  run_tags?: { [key: string]: any }
  cpu?: number
  memory?: number
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
