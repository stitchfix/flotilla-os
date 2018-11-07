import React, { Component } from "react"
import PropTypes from "prop-types"
import update from "immutability-helper"
import Ansi from "ansi-to-react"
import { has } from "lodash"
import axios from "axios"
import { ChevronUp, ChevronDown } from "react-feather"
import Button from "./Button"
import Card from "./Card"
import EmptyTable from "./EmptyTable"
import Loader from "./Loader"
import runStatusTypes from "../constants/runStatusTypes"
import config from "../config"

export default class RunLogs extends Component {
  static propTypes = {
    runId: PropTypes.string.isRequired,
    status: PropTypes.oneOf(Object.values(runStatusTypes)),
  }
  constructor(props) {
    super(props)
    this.scrollToBottom = this.scrollToBottom.bind(this)
    this.scrollToTop = this.scrollToTop.bind(this)
    this.handleAutoscrollChange = this.handleAutoscrollChange.bind(this)
  }
  state = {
    isLoading: false,
    error: false,
    lastSeen: undefined,
    logs: [],
    shouldAutoscroll: true,
  }
  componentDidMount() {
    this.fetch(this.props.runId)
    this.startInterval()
  }
  componentWillReceiveProps(nextProps) {
    if (this.props.runId !== nextProps.runId) {
      this.stopInterval()
      this.fetch(nextProps.runId)
      this.startInterval()
    } else if (nextProps.status === runStatusTypes.stopped) {
      this.stopInterval()
    }
  }
  componentDidUpdate(prevProps, prevState) {
    if (
      !!this.state.shouldAutoscroll &&
      this.state.logs.length > 0 &&
      this.state.logs.length !== prevState.logs.length
    ) {
      this.scrollToBottom()
    }
  }
  shouldComponentUpdate(nextProps, nextState) {
    // Compare loading and error states.
    if (
      this.state.isLoading !== nextState.isLoading ||
      this.state.error !== nextState.error ||
      this.state.shouldAutoscroll !== nextState.shouldAutoscroll
    ) {
      return true
    }

    // If loading and error states are equal, but the logs haven't changed,
    // don't update.
    if (this.state.logs.length === nextState.logs.length) {
      return false
    }
    return true
  }
  componentWillUnmount() {
    this.stopInterval()
  }
  fetch(runId) {
    const { lastSeen } = this.state
    let url = `${config.FLOTILLA_API}/${runId}/logs`

    // Append last_seen parameter if necessary.
    if (!!lastSeen) {
      url += `?last_seen=${lastSeen}`
    }

    return axios.get(url).then(({ data }) => {
      if (!!data.error) {
        this.stopInterval()
        this.setState({ error: data.error, isLoading: false })
      } else {
        if (!(data.last_seen === lastSeen && data.log === "")) {
          const logsArray = data.log.split("\n")
          this.setState({
            lastSeen: data.last_seen,
            logs: update(this.state.logs, { $push: logsArray }),
          })
        }

        if (!lastSeen || data.last_seen !== lastSeen) {
          if (has(data, "last_seen")) {
            this.fetch(this.props.runId)
          }
        }
      }
    })
  }
  startInterval() {
    this.setState({ isLoading: true })
    this._logsInterval = window.setInterval(() => {
      this.fetch(this.props.runId)
    }, 5000)
  }
  stopInterval() {
    this.setState({ isLoading: false })
    window.clearInterval(this._logsInterval)
  }
  handleAutoscrollChange(evt) {
    this.setState(state => ({ shouldAutoscroll: !state.shouldAutoscroll }))
  }
  scrollToBottom() {
    this.logsContainer.scrollTop = this.logsContainer.scrollHeight
  }
  scrollToTop() {
    this.logsContainer.scrollTop = 0
  }
  render() {
    const { shouldAutoscroll, error, isLoading, logs } = this.state
    const loaderContainerHeight = 50
    let content

    if (error) {
      content = <div>{error}</div>
    } else if (logs.length > 0) {
      content = (
        <pre
          className="flot-logs-container"
          ref={logsContainer => {
            this.logsContainer = logsContainer
          }}
        >
          {logs.map((line, i) => <Ansi key={i}>{line}</Ansi>)}
          {!!isLoading && <Loader />}
        </pre>
      )
    } else if (logs.length === 0) {
      content = <EmptyTable title="No logs yet!" />
    }

    return (
      <Card
        containerStyle={{ width: "100%" }}
        header={
          <div className="flex ff-rn j-sb a-c full-width">
            <div>Logs</div>
            <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
              <Button onClick={this.scrollToBottom}>
                <ChevronDown size={14} />
              </Button>
              <Button onClick={this.scrollToTop}>
                <ChevronUp size={14} />
              </Button>
              <div className="flex with-horizontal-child-margin">
                <input
                  type="checkbox"
                  onChange={this.handleAutoscrollChange}
                  checked={shouldAutoscroll}
                />
                <div>Autoscroll</div>
              </div>
            </div>
          </div>
        }
      >
        {content}
      </Card>
    )
  }
}
