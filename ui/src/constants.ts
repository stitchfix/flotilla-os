import { EnhancedRunStatus, RunStatus } from "./types"
import { Colors } from "@blueprintjs/core"
import { ReactJsonViewProps } from "react-json-view"

export const PAGE_SIZE = 20
export const RUN_FETCH_INTERVAL_MS = 5000 // 5 sec
export const LOG_FETCH_INTERVAL_MS = 10000 // 10 sec
export const KILL_LOG_POLLING_TIMEOUT_MS = 120000 // 2 mins
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
export const CHAR_TO_PX_RATIO = 40 / 300
export const JSON_VIEW_PROPS: Partial<ReactJsonViewProps> = {
  name: false,
  collapsed: 2,
  enableClipboard: false,
  displayDataTypes: false,
  displayObjectSize: false,
  theme: "ocean",
  style: {
    background: Colors.DARK_GRAY1,
    fontFamily: "Roboto Mono",
    fontSize: "0.8rem",
  },
}
