import React, { Component } from "react"
import PropTypes from "prop-types"
import { has, isEmpty } from "lodash"
import runStatusTypes from "../../helpers/runStatusTypes"
import api from "../../api"
import LogChunk from "./LogChunk"
import LogProcessor from "./LogProcessor"
import config from "../../config"

class LogRequester extends Component {
  state = {
    logs: [],
    lastSeen: null,
    inFlight: false,
    error: false,
  }

  componentDidMount() {
    this.requestLogs()

    if (this.props.status !== runStatusTypes.stopped) {
      this.setRequestInterval()
    }
  }

  componentDidUpdate(prevProps) {
    if (
      prevProps.status !== runStatusTypes.stopped &&
      this.props.status === runStatusTypes.stopped
    ) {
      this.clearRequestInterval()
    }
  }

  componentWillUnmount() {
    this.clearRequestInterval()
  }

  setRequestInterval = () => {
    this.requestInterval = window.setInterval(
      this.requestLogs,
      config.RUN_LOGS_REQUEST_INTERVAL_MS
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
  handleResponse = async response => {
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
  shouldAppendLogsToState = response =>
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
  appendLogsToState = response =>
    new Promise((resolve, reject) => {
      // Create a new LogChunk object.
      const chunk = new LogChunk({
        chunk: response.log,
        lastSeen: response.last_seen,
      })

      let prevLastSeen

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

  hasAdditionalLogs = ({ prevLastSeen, response }) => {
    if (!prevLastSeen || response.last_seen !== prevLastSeen) {
      if (has(response, "last_seen")) {
        return true
      }
    }

    return false
  }

  hasRunFinished = () => this.props.status === runStatusTypes.stopped

  render() {
    return <LogProcessor logs={this.state.logs} status={this.props.status} />
  }
}

LogRequester.propTypes = {
  runID: PropTypes.string,
  status: PropTypes.oneOf(Object.values(runStatusTypes)),
}

export default LogRequester
