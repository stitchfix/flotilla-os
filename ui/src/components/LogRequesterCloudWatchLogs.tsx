import * as React from "react"
import { has, isEmpty } from "lodash"
import api from "../api"
import Log from "./Log"
import { RunStatus, RunLog } from "../types"
import { LOG_FETCH_INTERVAL_MS } from "../constants"
import ErrorCallout from "./ErrorCallout"

type Props = {
  status: RunStatus | undefined
  runID: string
  height: number
  setHasLogs: () => void
  shouldAutoscroll: boolean
}

type State = {
  logs: string
  lastSeen: string | undefined
  isLoading: boolean
  error: any
  hasLogs: boolean
}

const initialState: State = {
  logs: "",
  lastSeen: undefined,
  isLoading: false,
  error: false,
  hasLogs: false,
}

class LogRequesterCloudWatchLogs extends React.Component<Props, State> {
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

    // Stop request interval if run transitions from running to stopped.
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

  /**
   * Performs one initial API call to the logs endpoint and starts a request
   * interval if the run is not stopped.
   */
  initialize() {
    this.requestLogs()

    if (this.props.status !== RunStatus.STOPPED) {
      this.setRequestInterval()
    }
  }

  /**
   * Clears the request interval, resets the component state, and calls
   * this.initialize.
   */
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
      .then((res: RunLog) => {
        this.handleResponse(res)
      })
      .catch(error => {
        this.clearRequestInterval()
        this.setState({ isLoading: false, error })
      })
  }

  handleResponse = (res: RunLog) => {
    const PREV_LAST_SEEN = this.state.lastSeen
    this.setState(
      prev => {
        const isLoading = false
        const error = false

        // Return if there are no logs.
        if (!has(res, "log") || isEmpty(res.log)) {
          return { ...prev, isLoading, error }
        }

        let logs = prev.logs
        let lastSeen: string | undefined = res.last_seen
        let hasLogs = prev.hasLogs || res.log.length > 0

        // Append logs if necessary.
        if (res.last_seen && res.last_seen !== prev.lastSeen) {
          logs += res.log
        }

        return { ...prev, isLoading, error, logs, lastSeen, hasLogs }
      },
      () => {
        if (
          this.props.status === RunStatus.STOPPED &&
          (!PREV_LAST_SEEN || res.last_seen !== PREV_LAST_SEEN)
        ) {
          if (has(res, "last_seen")) {
            this.requestLogs()
          }
        }
      }
    )
  }

  render() {
    const { height, status, shouldAutoscroll } = this.props
    const { isLoading, error, logs } = this.state

    if (error) return <ErrorCallout error={error} />

    return (
      <Log
        height={height}
        logs={logs}
        hasRunFinished={status === RunStatus.STOPPED}
        isLoading={isLoading}
        shouldAutoscroll={shouldAutoscroll}
      />
    )
  }
}

export default LogRequesterCloudWatchLogs
