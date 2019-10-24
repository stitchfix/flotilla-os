import { Run, RunStatus, EnhancedRunStatus } from "../types"

const getEnhancedRunStatus = (run: Run): EnhancedRunStatus | RunStatus => {
  if (run.status === RunStatus.STOPPED) {
    if (run.exit_code === 0) {
      return EnhancedRunStatus.SUCCESS
    } else {
      return EnhancedRunStatus.FAILED
    }
  }

  return run.status
}

export default getEnhancedRunStatus
