import * as React from "react"
import { Spinner, Pre, Colors, Classes } from "@blueprintjs/core"
import { has, isEmpty } from "lodash"
import { AxiosError } from "axios"
import { RequestStatus } from "./Request"
import { RunLog, RunStatus } from "../types"
import { LOG_FETCH_INTERVAL_MS } from "../constants"

export type Props = {
  runID: string
  status: RunStatus
  requestFn: ({
    runID,
    lastSeen,
  }: {
    runID: string
    lastSeen?: string
  }) => Promise<RunLog>
}
export type State = {
  requestStatus: RequestStatus
  data: RunLog[]
  isLoading: boolean
  error: AxiosError | null
  lastSeen: string | undefined
  totalLogsLength: number
}

class RunLogs extends React.Component<Props, State> {
  requestInterval: number | undefined
  state = {
    requestStatus: RequestStatus.NOT_READY,
    data: [],
    isLoading: false,
    error: null,
    lastSeen: undefined,
    totalLogsLength: 0,
  }

  componentDidMount = () => {
    this.request()

    if (this.props.status !== RunStatus.STOPPED) {
      this.setRequestInterval()
    }
  }

  componentDidUpdate = (prevProps: Props) => {
    if (
      prevProps.status !== RunStatus.STOPPED &&
      this.props.status === RunStatus.STOPPED
    ) {
      this.clearRequestInterval()
      this.request()
    }
  }

  componentWillUnmount = () => {
    this.clearRequestInterval()
  }

  setRequestInterval = () => {
    this.requestInterval = window.setInterval(
      this.request,
      LOG_FETCH_INTERVAL_MS
    )
  }

  clearRequestInterval = () => {
    window.clearInterval(this.requestInterval)
    this.requestInterval = undefined
  }

  request = () => {
    // Return if the request is in flight or if there's an error
    if (this.state.isLoading === true) return
    if (this.state.error !== null) return

    this.setState({ isLoading: true })
    this.props
      .requestFn({ runID: this.props.runID, lastSeen: this.state.lastSeen })
      .then(this.handleResponse)
      .catch(this.handleError)
  }

  handleResponse = async (response: RunLog) => {
    this.setState({
      isLoading: false,
      error: null,
      requestStatus: RequestStatus.READY,
    })

    // Return if there are no logs.
    if (!has(response, "log") || isEmpty(response.log)) {
      return
    }

    const shouldAppendLogsToState = this.shouldAppendLogsToState(response)

    if (shouldAppendLogsToState === true) {
      this.appendLogsToState(response)
        .then(prevLastSeen => {
          if (
            this.hasRunFinished() &&
            this.hasAdditionalLogs({ prevLastSeen, response })
          ) {
            this.request()
          }
        })
        .catch(err => {
          console.warn(err)
        })
    }
  }

  shouldAppendLogsToState = (response: RunLog): boolean =>
    response.last_seen === this.state.lastSeen ? false : true

  appendLogsToState = (response: RunLog): Promise<string | undefined> => {
    return new Promise(resolve => {
      let prevLastSeen: string | undefined

      // Append it to state.logs
      this.setState(
        prevState => {
          prevLastSeen = prevState.lastSeen

          return {
            data: [...prevState.data, response],
            lastSeen: response.last_seen,
            totalLogsLength: prevState.totalLogsLength + response.log.length,
          }
        },
        () => {
          resolve(prevLastSeen)
        }
      )
    })
  }

  hasAdditionalLogs = ({
    prevLastSeen,
    response,
  }: {
    prevLastSeen: string | undefined
    response: RunLog
  }): boolean => {
    if (!prevLastSeen || response.last_seen !== prevLastSeen) {
      if (has(response, "last_seen")) return true
    }

    return false
  }

  hasRunFinished = (): boolean => this.props.status === RunStatus.STOPPED

  handleError = (error: any) => {
    this.setState({ error })
  }

  render() {
    const { requestStatus, data } = this.state

    if (requestStatus === RequestStatus.READY && data) {
      return (
        <div className="flotilla-logs-container">
          {data.map((chunk: RunLog) => (
            <Pre
              className={`flotilla-pre ${Classes.DARK}`}
              key={chunk.last_seen}
              style={{
                opacity: 1,
                background: Colors.DARK_GRAY2,
                overflowY: "scroll",
                whiteSpace: "pre-wrap",
              }}
            >
              {chunk.log}
            </Pre>
          ))}
        </div>
      )
    }

    if (requestStatus === RequestStatus.ERROR) return <div>errro</div>
    return <Spinner />
  }
}

export default RunLogs
