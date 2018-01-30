import React, { Component } from "react"
import PropTypes from "prop-types"
import update from "immutability-helper"
import Ansi from "ansi-to-react"
import { has, get } from "lodash"
import { List, AutoSizer, CellMeasurer } from "react-virtualized"
import axios from "axios"
import { Card, Loader } from "aa-ui-components"
import config from "../config"
import { runStatusTypes } from "../constants/"
import EmptyTable from "./EmptyTable"

// Constants for React Virtualized to calculate row height based on number of
// chars per line.
const rowHeight = 26

// Estimated char width.
const estCharWidth = 7.645

// Max number of chars allowed per row, calculated by dividing the `width`
// from <Autosizer> by the estimated char width.
const maxCharsPerRow = width => width / estCharWidth

const rowStyles = {
  whiteSpace: "pre-wrap",
  wordBreak: "break-all",
  lineHeight: 1.4,
}

export default class RunLogs extends Component {
  static propTypes = {
    runId: PropTypes.string.isRequired,
    status: PropTypes.oneOf(Object.values(runStatusTypes)),
  }
  constructor(props) {
    super(props)
    this.rowRenderer = this.rowRenderer.bind(this)
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
    if (nextProps.status === runStatusTypes.stopped) {
      this.stopInterval()
    }
    if (this.props.runId !== nextProps.runId) {
      this.stopInterval()
      this.fetch(nextProps.runId)
      this.startInterval()
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
    // Don't fetch if the run is pending or queued.
    if (
      this.props.status === runStatusTypes.queued ||
      this.props.status === runStatusTypes.pending
    ) {
      return false
    }

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
  rowRenderer({ index, key, style, isScrolling }) {
    const { logs, isLoading } = this.state

    if (index === logs.length && !!isLoading) {
      return (
        <div key={key} style={style}>
          <Loader />
        </div>
      )
    }

    return (
      <div key={key} style={{ ...style, ...rowStyles }}>
        <Ansi>{logs[index]}</Ansi>
      </div>
    )
  }
  getVirtualizedHeight() {
    const topbarHeight = 48
    const viewHeaderHeight = 80
    const viewHeaderMarginBottom = 24
    const contentMarginBottom = 24
    const viewInnerMarginBottom = 72
    const runStatusBarHeight = 65 + 24 // height + margin-bottom

    return (
      window.innerHeight -
      topbarHeight -
      viewHeaderHeight -
      viewHeaderMarginBottom -
      contentMarginBottom -
      viewInnerMarginBottom -
      runStatusBarHeight
    )
  }
  handleAutoscrollChange(evt) {
    this.setState(state => ({ shouldAutoscroll: !state.shouldAutoscroll }))
  }
  render() {
    const { shouldAutoscroll, error, isLoading, logs } = this.state
    const loaderContainerHeight = 50
    let content

    if (error) {
      content = <div>{error}</div>
    } else if (logs.length > 0) {
      content = (
        <div className="full-width" style={{ height: "100%" }}>
          <AutoSizer disableHeight>
            {({ width }) => {
              const rowCount = !!isLoading ? logs.length + 1 : logs.length

              let scrollToIndex

              if (shouldAutoscroll) {
                scrollToIndex = !!isLoading ? logs.length : logs.length - 1
              }
              return (
                <List
                  className="code"
                  width={width}
                  height={this.getVirtualizedHeight()}
                  rowCount={rowCount}
                  rowRenderer={this.rowRenderer}
                  rowHeight={({ index }) => {
                    if (index === logs.length) {
                      return loaderContainerHeight
                    }

                    if (logs[index].length <= maxCharsPerRow(width)) {
                      return rowHeight
                    } else {
                      return (
                        rowHeight * (logs[index].length / maxCharsPerRow(width))
                      )
                    }
                  }}
                  scrollToIndex={scrollToIndex}
                />
              )
            }}
          </AutoSizer>
        </div>
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
              <div className="pl-button with-horizontal-child-margin">
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
