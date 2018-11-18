import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import { capitalize } from "lodash"
import { CheckCircle, XCircle } from "react-feather"
import colors from "../../constants/colors"
import runStatusTypes from "../../constants/runStatusTypes"
import Loader from "../styled/Loader"

const RunStatusContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: flex-end;

  & > :first-child {
    margin-right: 4px;
  }
`

const getHumanReadableStatus = ({ status, exitCode }) => {
  if (status === runStatusTypes.stopped) {
    if (exitCode === 0) {
      return runStatusTypes.success
    } else {
      return runStatusTypes.failed
    }
  }
  return status
}

const getIconByStatus = status => {
  switch (status) {
    case runStatusTypes.queued:
      return <Loader spinnerStyle={{ borderLeftColor: colors.gray[4] }} />
    case runStatusTypes.pending:
      return <Loader spinnerStyle={{ borderLeftColor: colors.yellow[0] }} />
    case runStatusTypes.running:
      return <Loader />
    case runStatusTypes.success:
      return <CheckCircle size={14} color={colors.green[0]} />
    case runStatusTypes.failed:
    case runStatusTypes.needs_retry:
      return <XCircle size={14} color={colors.red[0]} />
    default:
      return null
  }
}

const RunStatus = ({ exitCode, onlyRenderIcon, status }) => {
  const readableStatus = getHumanReadableStatus({ status, exitCode })
  const icon = getIconByStatus(readableStatus)

  if (onlyRenderIcon) {
    return icon
  }

  return (
    <RunStatusContainer>
      <h3>{capitalize(readableStatus)}</h3>
      <div>{icon}</div>
    </RunStatusContainer>
  )
}

RunStatus.displayName = "RunStatus"

RunStatus.propTypes = {
  exitCode: PropTypes.number,
  onlyRenderIcon: PropTypes.bool.isRequired,
  status: PropTypes.oneOf(Object.values(runStatusTypes)),
}

RunStatus.defaultProps = {
  onlyRenderIcon: true,
}

export default RunStatus
