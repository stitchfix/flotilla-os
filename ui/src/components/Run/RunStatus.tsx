import * as React from "react"
import styled from "styled-components"
import { capitalize } from "lodash"
import { CheckCircle, XCircle } from "react-feather"
import colors from "../../helpers/colors"
import Loader from "../styled/Loader"
import { ecsRunStatuses, intents } from "../../.."

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
  status: ecsRunStatuses
  exitCode?: number
}) => {
  if (status === ecsRunStatuses.STOPPED) {
    if (exitCode === 0) {
      return ecsRunStatuses.SUCCESS
    } else {
      return ecsRunStatuses.FAILED
    }
  }
  return status
}

interface IRunStatusProps {
  exitCode?: number
  onlyRenderIcon: boolean
  status: ecsRunStatuses
}

class RunStatus extends React.PureComponent<IRunStatusProps> {
  static displayName = "RunStatus"

  static defaultProps: Partial<IRunStatusProps> = {
    onlyRenderIcon: true,
  }

  getIconByStatus = (status: ecsRunStatuses): React.ReactNode => {
    switch (status) {
      case ecsRunStatuses.QUEUED:
        return <Loader intent={intents.SUBTLE} />
      case ecsRunStatuses.PENDING:
        return <Loader intent={intents.WARNING} />
      case ecsRunStatuses.RUNNING:
        return <Loader intent={intents.PRIMARY} />
      case ecsRunStatuses.SUCCESS:
        return <CheckCircle size={14} color={colors.green[0]} />
      case ecsRunStatuses.FAILED:
      case ecsRunStatuses.NEEDS_RETRY:
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
