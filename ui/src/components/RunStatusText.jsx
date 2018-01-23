import React from 'react'
import {
  CheckCircle,
  XCircle,
} from 'react-feather'
import { capitalize } from 'lodash'
import { runStatusTypes } from '../constants/'
import { getRunStatus } from '../utils/'
import { Loader } from './'

export default function RunStatusText({
  status,
  exitCode,
  enhancedStatus,
}) {
  // Generates a human readable status; differentiates between success
  // and failed runs, both of which have a status of stopped.
  const _enhancedStatus = !!enhancedStatus ? enhancedStatus :
    getRunStatus({ status, exitCode })
  let icon

  switch (_enhancedStatus) {
    case runStatusTypes.queued:
      icon = (
        <Loader mini />
      )
      break
    case runStatusTypes.pending:
      icon = (
        <Loader mini />
      )
      break
    case runStatusTypes.running:
      icon = (
        <Loader mini />
      )
      break
    case runStatusTypes.success:
      icon = <CheckCircle size={14} />
      break
    case runStatusTypes.failed:
      icon = <XCircle size={14} />
      break
    case runStatusTypes.needs_retry:
      icon = <XCircle size={14} />
      break
    default:
      icon = null
  }
  return (
    <div className={`run-status-text-container flex ${enhancedStatus}`}>
      <div className="run-status-text">{capitalize(enhancedStatus)}</div>
      <div className="run-status-icon">{icon}</div>
    </div>
  )
}
