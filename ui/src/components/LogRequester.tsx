import * as React from "react"
import { has, isEmpty } from "lodash"
import api from "../api"
import LogRenderer from "./LogRenderer"
import { LogChunk, RunStatus, RunLog } from "../types"
import { LOG_FETCH_INTERVAL_MS } from "../constants"
import ErrorCallout from "./ErrorCallout"

type Props = {
  status: RunStatus | undefined
  runID: string
  height: number
  setHasLogs: () => void
}

type State = {
  logs: LogChunk[]
  lastSeen: string | undefined
  isLoading: boolean
  error: any
  hasLogs: boolean
}

const initialState: State = {
  logs: [],
  lastSeen: undefined,
  isLoading: false,
  error: false,
  hasLogs: false,
}

class LogRequester extends React.PureComponent<Props, State> {
  private requestInterval: number | undefined
  state = initialState

  componentDidMount() {
    this.initialize()
  }

  componentDidUpdate(prevProps: Props, prevState: State) {
    if (prevProps.runID !== this.props.runID) {
      this.handleRunIDChange()
      return
    }

    if (
      prevProps.status !== RunStatus.STOPPED &&
      this.props.status === RunStatus.STOPPED
    ) {
      this.clearRequestInterval()
    }

    if (prevState.hasLogs === false && this.state.hasLogs === true) {
      this.props.setHasLogs()
    }
  }

  componentWillUnmount() {
    this.clearRequestInterval()
  }

  setRequestInterval = (): void => {
    this.requestInterval = window.setInterval(
      this.requestLogs,
      LOG_FETCH_INTERVAL_MS
    )
  }

  clearRequestInterval = () => {
    window.clearInterval(this.requestInterval)
  }

  initialize() {
    this.requestLogs()

    if (this.props.status !== RunStatus.STOPPED) {
      this.setRequestInterval()
    }
  }

  handleRunIDChange() {
    // Clear request interval
    this.clearRequestInterval()

    // Reset state.
    this.setState(initialState, () => {
      // Initialize, as if the component just mounted.
      this.initialize()
    })
  }

  requestLogs = () => {
    const { runID } = this.props
    const { lastSeen } = this.state

    this.setState({ isLoading: true })

    api
      .getRunLog({ runID, lastSeen })
      .then(this.handleResponse)
      .catch(error => {
        this.clearRequestInterval()
        this.setState({ error })
      })
  }

  /**
   * The response handler for the logs endpoint performs the following:
   * 1. Calls this.shouldAppendLogsToState to determine if a valid log chunk
   * was received.
   * 2. Once the logs have been appended, it checks to see if it should
   * "exhaust" the logs endpoint - meaning that it will continuously hit the
   * logs endpoint until all remaining logs have been fetched.
   */
  handleResponse = async (response: RunLog) => {
    this.setState({ isLoading: false, error: false })

    // Return if there are no logs.
    if (!has(response, "log") || isEmpty(response.log)) {
      return
    }

    const shouldAppendLogsToState = await this.shouldAppendLogsToState(response)

    if (shouldAppendLogsToState === true) {
      this.appendLogsToState(response).then(prevLastSeen => {
        if (
          this.hasRunFinished() &&
          this.hasAdditionalLogs({ prevLastSeen, response })
        ) {
          this.requestLogs()
        }
      })
    }
  }

  /**
   * Returns a Promise that resolves with whether logs should be appended to
   * the component's state.
   *
   * Note: this probably should just return a boolean but I haven't had time
   * to test what that would look like on long running jobs.
   */
  shouldAppendLogsToState = (response: RunLog): Promise<boolean> =>
    new Promise(resolve => {
      const { lastSeen } = this.state

      if (response.last_seen === lastSeen) {
        resolve(false)
      }

      resolve(true)
    })

  /**
   * Returns a Promise that appends a new LogChunk object to the component's
   * state and resolves with the previous state's `lastSeen` attribute. The
   * previous state's lastSeen is then used to determine whether or not there
   * are additional logs still waiting to be fetched.
   */
  appendLogsToState = (response: RunLog): Promise<string | undefined> =>
    new Promise((resolve, reject) => {
      // Create a new LogChunk object.
      const chunk: LogChunk = {
        chunk: response.log,
        lastSeen: response.last_seen,
      }

      let prevLastSeen: string | undefined

      // Append it to state.logs
      this.setState(
        prevState => {
          const prevHasLogs = prevState.hasLogs
          prevLastSeen = prevState.lastSeen
          return {
            logs: [...prevState.logs, chunk],
            lastSeen: response.last_seen,
            hasLogs: prevHasLogs === true ? true : response.log.length > 0,
          }
        },
        () => {
          resolve(prevLastSeen)
        }
      )
    })

  hasAdditionalLogs = ({
    prevLastSeen,
    response,
  }: {
    prevLastSeen: string | undefined
    response: RunLog
  }): boolean => {
    if (!prevLastSeen || response.last_seen !== prevLastSeen) {
      if (has(response, "last_seen")) {
        return true
      }
    }

    return false
  }

  hasRunFinished = (): boolean => this.props.status === RunStatus.STOPPED

  render() {
    const { height } = this.props
    const { logs, isLoading, error } = this.state

    if (error) return <ErrorCallout error={error} />

    return (
      <LogRenderer
        height={height}
        logs={logs}
        hasRunFinished={this.hasRunFinished()}
        isLoading={isLoading}
      />
    )
  }
}

export default LogRequester
