import React, { Component } from "react"
import PropTypes from "prop-types"
import { get, has, isEmpty } from "lodash"
import runStatusTypes from "../../constants/runStatusTypes"
import api from "../../api"
import LogChunk from "./LogChunk"
import LogRenderer from "./LogRenderer"
import config from "../../config"

class LogRequester extends Component {
  state = {
    logs: [],
    lastSeen: null,
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

    api
      .getRunLogs({ runID, lastSeen })
      .then(this.handleResponse)
      .catch(error => {
        console.log(error)
      })
  }

  handleResponse = async response => {
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
    return <LogRenderer logs={this.state.logs} />
  }
}

LogRequester.propTypes = {
  runID: PropTypes.string,
  status: PropTypes.oneOf(Object.values(runStatusTypes)),
}

export default LogRequester
