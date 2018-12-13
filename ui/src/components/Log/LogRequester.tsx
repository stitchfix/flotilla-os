import * as React from "react"
import { has, isEmpty } from "lodash"
import api from "../../api"
import LogProcessor from "./LogProcessor"
import config from "../../config"
import {
  IFlotillaUILogChunk,
  ecsRunStatuses,
  IFlotillaAPILogsResponse,
} from "../../.."

interface ILogRequesterProps {
  status: ecsRunStatuses | undefined
  runID: string
}

interface ILogRequesterState {
  logs: IFlotillaUILogChunk[]
  lastSeen: string | undefined
  inFlight: boolean
  error: any
}

class LogRequester extends React.PureComponent<
  ILogRequesterProps,
  ILogRequesterState
> {
  private requestInterval: number | undefined

  state = {
    logs: [],
    lastSeen: undefined,
    inFlight: false,
    error: false,
  }

  componentDidMount() {
    this.requestLogs()

    if (this.props.status !== ecsRunStatuses.STOPPED) {
      this.setRequestInterval()
    }
  }

  componentDidUpdate(prevProps: ILogRequesterProps) {
    if (
      prevProps.status !== ecsRunStatuses.STOPPED &&
      this.props.status === ecsRunStatuses.STOPPED
    ) {
      this.clearRequestInterval()
    }
  }

  componentWillUnmount() {
    this.clearRequestInterval()
  }

  setRequestInterval = (): void => {
    this.requestInterval = window.setInterval(
      this.requestLogs,
      +config.RUN_LOGS_REQUEST_INTERVAL_MS
    )
  }

  clearRequestInterval = () => {
    window.clearInterval(this.requestInterval)
  }

  requestLogs = () => {
    const { runID } = this.props
    const { lastSeen } = this.state

    this.setState({ inFlight: true })

    api
      .getRunLogs({ runID, lastSeen })
      .then(this.handleResponse)
      .catch(error => {
        this.clearRequestInterval()
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
  handleResponse = async (response: IFlotillaAPILogsResponse) => {
    this.setState({ inFlight: false, error: false })

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
   */
  shouldAppendLogsToState = (
    response: IFlotillaAPILogsResponse
  ): Promise<boolean> =>
    new Promise((resolve, reject) => {
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
  appendLogsToState = (
    response: IFlotillaAPILogsResponse
  ): Promise<string | undefined> =>
    new Promise((resolve, reject) => {
      // Create a new LogChunk object.
      const chunk: IFlotillaUILogChunk = {
        chunk: response.log,
        lastSeen: response.last_seen,
      }

      let prevLastSeen: string | undefined

      // Append it to state.logs
      this.setState(
        prevState => {
          prevLastSeen = prevState.lastSeen
          return {
            logs: [...prevState.logs, chunk],
            lastSeen: response.last_seen,
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
    response: IFlotillaAPILogsResponse
  }): boolean => {
    if (!prevLastSeen || response.last_seen !== prevLastSeen) {
      if (has(response, "last_seen")) {
        return true
      }
    }

    return false
  }

  hasRunFinished = (): boolean => this.props.status === ecsRunStatuses.STOPPED

  render() {
    return <LogProcessor logs={this.state.logs} />
  }
}

export default LogRequester
