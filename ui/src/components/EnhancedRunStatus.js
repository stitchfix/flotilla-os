import React from "react"
import { CheckCircle, XCircle } from "react-feather"
import { capitalize } from "lodash"
import colors from "../constants/colors"
import runStatusTypes from "../constants/runStatusTypes"
import Loader from "./Loader"

export const getEnhancedStatus = (status, exitCode) => {
  if (status === runStatusTypes.stopped) {
    if (exitCode === 0) {
      return runStatusTypes.success
    } else {
      return runStatusTypes.failed
    }
  }
  return status
}
export const getIcon = enhancedStatus => {
  switch (enhancedStatus) {
    case runStatusTypes.queued:
      return (
        <Loader mini spinnerStyle={{ borderLeftColor: colors.gray.gray_4 }} />
      )
    case runStatusTypes.pending:
      return (
        <Loader
          mini
          spinnerStyle={{ borderLeftColor: colors.yellow.yellow_0 }}
        />
      )
    case runStatusTypes.running:
      return <Loader mini />
    case runStatusTypes.success:
      return <CheckCircle size={14} color={colors.green.green_0} />
    case runStatusTypes.failed:
    case runStatusTypes.needs_retry:
      return <XCircle size={14} color={colors.red.red_0} />
    default:
      return null
  }
}

export default function EnhancedRunStatus({ status, exitCode, iconOnly }) {
  const enhancedStatus = getEnhancedStatus(status, exitCode)
  const icon = getIcon(enhancedStatus)
  const className = `run-status-text-container flex ff-rn j-fs a-fe ${enhancedStatus}`

  if (iconOnly) {
    return <div className="run-status-icon">{icon}</div>
  }

  return (
    <div className={className} style={{ fontSize: "1rem", fontWeight: 400 }}>
      <div className="run-status-text">{capitalize(enhancedStatus)}</div>
      <div className="run-status-icon" style={{ marginLeft: 4 }}>
        {icon}
      </div>
    </div>
  )
}
