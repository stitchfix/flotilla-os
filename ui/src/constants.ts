import { EnhancedRunStatus, RunStatus } from "./types"
import { Colors } from "@blueprintjs/core"

export const PAGE_SIZE = 20
export const RUN_FETCH_INTERVAL_MS = 5000 // 5 sec
export const LOG_FETCH_INTERVAL_MS = 10000 // 10 sec
export const RUN_TAB_ID_QUERY_KEY = "rt"
export const LOG_SEARCH_QUERY_KEY = "log_search"
export const RUN_STATUS_COLOR_MAP = new Map<
  EnhancedRunStatus | RunStatus,
  string
>([
  [EnhancedRunStatus.PENDING, Colors.GRAY3],
  [EnhancedRunStatus.QUEUED, Colors.GOLD5],
  [EnhancedRunStatus.RUNNING, Colors.COBALT4],
  [EnhancedRunStatus.STOPPED, Colors.RED4],
  [EnhancedRunStatus.NEEDS_RETRY, Colors.RED4],
  [EnhancedRunStatus.SUCCESS, Colors.GREEN5],
  [EnhancedRunStatus.FAILED, Colors.RED4],
])
export const LOCAL_STORAGE_SETTINGS_KEY = "settings"
export const LOCAL_STORAGE_IS_ONBOARDED_KEY = "is_onboarded"
