import { ReactNode } from "react"

export interface IFlotillaUITaskContext {
  data: IFlotillaTaskDefinition
  inFlight: boolean
  error: boolean
  requestState: requestStates
  definitionID: string
  requestData: () => void
}

export enum asyncDataTableFilters {
  INPUT = "INPUT",
  SELECT = "SELECT",
  CUSTOM = "CUSTOM",
  KV = "KV",
}

export interface IAsyncDataTableFilterProps {
  description?: string
  displayName: string
  field: string
  type: asyncDataTableFilters
  filterProps?: any
}

export interface IFlotillaUIConfig {
  DEFAULT_CLUSTER: string
  FLOTILLA_API: string
  IMAGE_PREFIX: string
  INVALID_RUN_ENV: string[]
  REQUIRED_RUN_TAGS: string[]
  RUN_LOGS_REQUEST_INTERVAL_MS: string | number
  RUN_REQUEST_INTERVAL_MS: string | number
}

export interface IFlotillaAPIError {
  data: any
  status?: any
  headers?: any
}

export interface IFlotillaEditTaskPayload {
  memory: number
  image: string
  group_name: string
  tags?: string[]
  env?: IFlotillaEnv[]
  command: string
}

/** The values required to create a task definition. */
export interface IFlotillaCreateTaskPayload extends IFlotillaEditTaskPayload {
  alias: string
}

export interface IFlotillaTaskDefinition {}

/** The values required to execute a task definition. */
export interface IFlotillaRunTaskPayload {}

export interface IFlotillaEnv {
  name: string
  value: any
}

export interface IFlotillaRun {
  status: ecsRunStatuses
  cluster: string
  finished_at?: string
  image: string
  run_id: string
  exit_code?: number
  group_name: string
  definition_id: string
  instance: {
    instance_id: string
    dns_name: string
  }
  alias: string
  env: IFlotillaEnv[]
  started_at?: string
}

export interface IFlotillaLogChunk {
  chunk: string
  lastSeen: string
}

export interface IReactSelectOption {
  label: string
  value: any
}

export interface IFlotillaUIBreadcrumb {
  text: string
  href: string
}

export interface IFlotillaUINavigationLink {
  isLink: boolean
  text: string
  href?: string
  buttonProps?: Partial<IFlotillaUIButtonProps>
}

export interface IFlotillaUIButtonProps
  extends React.HTMLProps<HTMLButtonElement> {
  intent?: intents
  isDisabled: boolean
  isLoading: boolean
  onClick?: (evt: React.SyntheticEvent) => void
  type: string
}

export enum intents {
  PRIMARY = "PRIMARY",
  SUCCESS = "SUCCESS",
  WARNING = "WARNING",
  ERROR = "ERROR",
  SUBTLE = "SUBTLE",
}

export enum requestStates {
  READY = "READY",
  NOT_READY = "NOT_READY",
  ERROR = "ERROR",
}

export enum ecsRunStatuses {
  PENDING = "PENDING",
  QUEUED = "QUEUED",
  RUNNING = "RUNNING",
  STOPPED = "STOPPED",
  NEEDS_RETRY = "NEEDS_RETRY",
  SUCCESS = "SUCCESS",
  FAILED = "FAILED",
}

export enum taskFormTypes {
  CREATE = "CREATE",
  EDIT = "EDIT",
  COPY = "COPY",
}

export interface IPopupProps {
  actions?: ReactNode
  body?: ReactNode
  intent?: intents
  shouldAutohide?: boolean
  title?: ReactNode
  unrenderPopup?: () => void
  visibleDuration?: number
}

export interface IPopupContext {
  renderPopup: (props: IPopupProps) => void
  unrenderPopup: () => void
}

export interface IModalContext {
  renderModal: (modal: React.ReactNode) => void
  unrenderModal: () => void
}
