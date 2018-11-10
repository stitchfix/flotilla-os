import React, { Component } from "react"
import PropTypes from "prop-types"
import { get, isString, isEmpty, has } from "lodash"
import * as requestStateTypes from "../../constants/requestStateTypes"
import api from "../../api"
import config from "../../config"
import runStatusTypes from "../../constants/runStatusTypes"

const LAST_SEEN_INITIAL_STATE = null

class RunLogs extends Component {
  state = {
    logs: [],
    lastSeen: LAST_SEEN_INITIAL_STATE,
  }

  componentDidMount() {
    this.requestLogs()

    this.requestInterval = window.setInterval(() => {
      this.requestLogs()
    }, config.RUN_LOGS_REQUEST_INTERVAL_MS)
  }

  componentDidUpdate(prevProps, prevState) {
    if (
      get(prevProps, "status") !== runStatusTypes.stopped &&
      get(this.props, "status") === runStatusTypes.stopped
    ) {
      this.clearInterval()
    }
  }

  componentWillUnmount() {
    window.clearInterval(this.requestInterval)
  }

  clearInterval = () => {
    window.clearInterval(this.requestInterval)
  }

  shouldAppendLogsToState = (log = "", lastSeen = LAST_SEEN_INITIAL_STATE) => {
    if (lastSeen === this.state.lastSeen && isEmpty(log)) {
      return false
    }

    return true
  }

  appendLogsToState = (log = "", lastSeen = LAST_SEEN_INITIAL_STATE) => {
    this.setState({
      logs: [...this.state.logs, log],
      lastSeen,
    })
  }

  hasMoreLogs = (lastSeen = LAST_SEEN_INITIAL_STATE) => {
    if (
      this.state.lastSeen === LAST_SEEN_INITIAL_STATE ||
      lastSeen !== this.state.lastSeen
    ) {
      if (lastSeen !== LAST_SEEN_INITIAL_STATE) {
        return true
      }
    }

    return false
  }

  requestLogs = () => {
    const { lastSeen } = this.state
    const { runID } = this.props

    api
      .getRunLogs({ runID, lastSeen })
      .then(data => {
        const l = get(data, "log")
        const ls = get(data, "last_seen", LAST_SEEN_INITIAL_STATE)

        if (this.shouldAppendLogsToState(l, ls)) {
          this.appendLogsToState(l, ls)
        }

        if (this.hasMoreLogs(ls)) {
          // @TODO
        }
      })
      .catch(error => {
        console.error(error)
      })
  }

  render() {
    const { logs } = this.state
    return (
      <div style={{ whiteSpace: "pre-line", fontFamily: "Source Code Pro" }}>
        {logs}
      </div>
    )
  }
}

RunLogs.propTypes = {
  runID: PropTypes.string.isRequired,
}

export default RunLogs
