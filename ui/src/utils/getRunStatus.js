import { runStatusTypes } from '../constants/'

// Generates a human-readable status for runs.
export default function getRunStatus({ status, exitCode }) {
  let _status = ''
  if (!!status) {
    if (status.toLowerCase() === runStatusTypes.stopped) {
      if (exitCode === 0) {
        _status = runStatusTypes.success
      } else {
        _status = runStatusTypes.failed
      }
    } else if (status.toLowerCase() === runStatusTypes.running) {
      _status = runStatusTypes.running
    } else if (status.toLowerCase() === runStatusTypes.pending) {
      _status = runStatusTypes.pending
    } else if (status.toLowerCase() === runStatusTypes.queued) {
      _status = runStatusTypes.queued
    }
  } else {
    _status = null
  }
  return _status
}
