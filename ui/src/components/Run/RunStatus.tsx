import * as React from "react"
import styled from "styled-components"
import { capitalize } from "lodash"
import { CheckCircle, XCircle } from "react-feather"
import colors from "../../helpers/colors"
import Loader from "../styled/Loader"
import { flotillaRunStatuses, flotillaUIIntents } from "../../types"

const RunStatusContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: flex-end;

  & > :first-child {
    margin-right: 4px;
  }
`

const getHumanReadableStatus = ({
  status,
  exitCode,
}: {
  status: flotillaRunStatuses
  exitCode?: number
}) => {
  if (status === flotillaRunStatuses.STOPPED) {
    if (exitCode === 0) {
      return flotillaRunStatuses.SUCCESS
    } else {
      return flotillaRunStatuses.FAILED
    }
  }
  return status
}

interface IRunStatusProps {
  exitCode?: number
  onlyRenderIcon: boolean
  status: flotillaRunStatuses
}

class RunStatus extends React.PureComponent<IRunStatusProps> {
  static displayName = "RunStatus"

  static defaultProps: Partial<IRunStatusProps> = {
    onlyRenderIcon: true,
  }

  getIconByStatus = (status: flotillaRunStatuses): React.ReactNode => {
    switch (status) {
      case flotillaRunStatuses.QUEUED:
        return <Loader intent={flotillaUIIntents.SUBTLE} />
      case flotillaRunStatuses.PENDING:
        return <Loader intent={flotillaUIIntents.WARNING} />
      case flotillaRunStatuses.RUNNING:
        return <Loader intent={flotillaUIIntents.PRIMARY} />
      case flotillaRunStatuses.SUCCESS:
        return <CheckCircle size={14} color={colors.green[0]} />
      case flotillaRunStatuses.FAILED:
      case flotillaRunStatuses.NEEDS_RETRY:
        return <XCircle size={14} color={colors.red[0]} />
      default:
        return null
    }
  }

  render() {
    const { exitCode, onlyRenderIcon, status } = this.props

    const readableStatus = getHumanReadableStatus({ status, exitCode })
    const icon = this.getIconByStatus(readableStatus)

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
}

export default RunStatus
