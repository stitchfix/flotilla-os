/**
 * This file contains various type definitions, interfaces, and enums. For
 * clarity, the naming conventions are grouped into 3 categories: UI-specific
 * (`IFlotillaUI...`), API-specific (`IFlotillaAPI...`), and shared (
 * `IFlotilla...`).
 */
import { ReactNode } from "react"
import { LocationDescriptor } from "history"

/** Error shape that the FlotillaAPIClient will reject with. */
export interface IFlotillaAPIError {
  data: any
  status?: any
  headers?: any
}

/** The response from the API when hitting the `/logs` endpoint. */
export interface IFlotillaAPILogsResponse {
  log: string
  last_seen: string
}

/** The config required to run and build the UI. */
export interface IFlotillaUIConfig {
  DEFAULT_CLUSTER: string
  FLOTILLA_API: string
  FLOTILLA_API_DEV: string
  IMAGE_PREFIX: string
  INVALID_RUN_ENV: string[]
  REQUIRED_RUN_TAGS: string[]
  RUN_LOGS_REQUEST_INTERVAL_MS: string | number
  RUN_REQUEST_INTERVAL_MS: string | number
}

/** The shape of the task React Context available to consumers. */
export interface IFlotillaUITaskContext {
  data: IFlotillaTaskDefinition | null
  inFlight: boolean
  error: boolean
  requestState: flotillaUIRequestStates
  definitionID: string
  requestData: () => void
}

/** The shape of the run React Context available to consumers. */
export interface IFlotillaUIRunContext {
  data: IFlotillaRun | null
  inFlight: boolean
  error: any
  requestState: flotillaUIRequestStates
  runID: string
}

/**
 * The UILogChunk wraps a slice of the ECS log output returned by Flotilla and
 * is used by the Log-related components (`src/components/Log/`).
 */
export interface IFlotillaUILogChunk {
  chunk: string
  lastSeen: string
}

/** Navigation breadcrumb shape. */
export interface IFlotillaUIBreadcrumb {
  text: string
  href: string
}

/** Navigation Link shape */
export interface IFlotillaUINavigationLink {
  isLink: boolean
  text: string
  href?: string | LocationDescriptor
  buttonProps?: Partial<IFlotillaUIButtonProps>
}

/** Props for styled button component. */
export interface IFlotillaUIButtonProps
  extends React.HTMLProps<HTMLButtonElement> {
  intent?: flotillaUIIntents
  isDisabled: boolean
  isLoading: boolean
  onClick?: (evt: React.SyntheticEvent) => void
  type: string
}

/** Task definition shared by API and UI. */
export interface IFlotillaTaskDefinition {
  alias: string
  arn: string
  command: string
  container_name: string
  definition_id: string
  env: IFlotillaEnv[]
  group_name: string
  image: string
  memory: number
  tags: string[]
}

/** Run information shared by API and UI. */
export interface IFlotillaRun {
  status: flotillaRunStatuses
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

/** Payload required to update a task definition. */
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

/** The values required to execute a task definition. */
export interface IFlotillaRunTaskPayload {
  run_tags?: IFlotillaEnv[]
  cluster: string
  env?: IFlotillaEnv[]
}

/** Flotilla environment variable, used in task definitions and execution. */
export interface IFlotillaEnv {
  name: string
  value: any
}

/** Filter types for the AsyncDataTable component */
export enum flotillaUIAsyncDataTableFilters {
  INPUT = "INPUT",
  SELECT = "SELECT",
  CUSTOM = "CUSTOM",
  KV = "KV",
}

/** Filter prop shape for AsyncDataTableFilter components. */
export interface IFlotillaUIAsyncDataTableFilterProps {
  description?: string
  displayName: string
  name: string
  type: flotillaUIAsyncDataTableFilters
  filterProps?: any
}

/** Shape of react-select option. */
export interface IReactSelectOption {
  label: string
  value: any
}

/**
 * Intents are used to indicate, say, what color a button is supposed to be.
 * See the `src/helpers/intentToColor.ts` helper to see what maps to what.
 */
export enum flotillaUIIntents {
  PRIMARY = "PRIMARY",
  SUCCESS = "SUCCESS",
  WARNING = "WARNING",
  ERROR = "ERROR",
  SUBTLE = "SUBTLE",
}

/** API request states. */
export enum flotillaUIRequestStates {
  READY = "READY",
  NOT_READY = "NOT_READY",
  ERROR = "ERROR",
}

/** Run statuses the API will return. */
export enum flotillaRunStatuses {
  PENDING = "PENDING",
  QUEUED = "QUEUED",
  RUNNING = "RUNNING",
  STOPPED = "STOPPED",
  NEEDS_RETRY = "NEEDS_RETRY",
  SUCCESS = "SUCCESS",
  FAILED = "FAILED",
}

/** The TaskForm component uses this to determine what to render. */
export enum flotillaUITaskFormTypes {
  CREATE = "CREATE",
  EDIT = "EDIT",
  COPY = "COPY",
}

/** Props for the Popup component. */
export interface IFlotillaUIPopupProps {
  actions?: ReactNode
  body?: ReactNode
  intent?: flotillaUIIntents
  shouldAutohide?: boolean
  title?: ReactNode
  unrenderPopup?: () => void
  visibleDuration?: number
}

/** Popup React Context available to consumers. */
export interface IFlotillaUIPopupContext {
  renderPopup: (props: IFlotillaUIPopupProps) => void
  unrenderPopup: () => void
}

/** Modal React Context available to consumers. */
export interface IFlotillaUIModalContext {
  renderModal: (modal: React.ReactNode) => void
  unrenderModal: () => void
}
