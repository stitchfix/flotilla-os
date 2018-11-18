import React, { Component, createRef, Fragment } from "react"
import PropTypes from "prop-types"
import { FixedSizeList as List } from "react-window"
import ReactResizeDetector from "react-resize-detector"
import { get, isEmpty, round } from "lodash"
import LogRow from "./LogRow"
import { RUN_BAR_HEIGHT_PX, NAVIGATION_HEIGHT_PX } from "../../constants/styles"
import RunBar from "./RunBar"
import RunContext from "./RunContext"
import runStatusTypes from "../../constants/runStatusTypes"

const LIST_REF = createRef()

class LogRenderer extends Component {
  state = {
    shouldAutoscroll: true,
  }

  componentDidMount() {
    LIST_REF.current.scrollToItem(this.props.len)
  }

  componentDidUpdate(prevProps) {
    if (
      this.state.shouldAutoscroll === true &&
      prevProps.len !== this.props.len
    ) {
      LIST_REF.current.scrollToItem(this.props.len)
    }
  }

  toggleShouldAutoscroll = () => {
    this.setState(prev => ({ shouldAutoscroll: !prev.shouldAutoscroll }))
  }

  handleScrollToTopClick = () => {
    LIST_REF.current.scrollToItem(0)
  }
  handleScrollToBottomClick = () => {
    LIST_REF.current.scrollToItem(this.props.len)
  }

  render() {
    const { width, height, logs, len } = this.props

    return (
      <RunContext.Consumer>
        {({ data }) => {
          const _len =
            get(data, "status") === runStatusTypes.stopped ? len : len + 1
          return (
            <Fragment>
              <RunBar
                shouldAutoscroll={this.state.shouldAutoscroll}
                toggleShouldAutoscroll={this.toggleShouldAutoscroll}
                onScrollToTopClick={this.handleScrollToTopClick}
                onScrollToBottomClick={this.handleScrollToBottomClick}
              />
              <List
                ref={LIST_REF}
                height={height - RUN_BAR_HEIGHT_PX}
                itemCount={_len}
                itemData={logs}
                itemSize={20}
                width={width}
                overscanCount={100}
                style={{ marginTop: RUN_BAR_HEIGHT_PX }}
              >
                {LogRow}
              </List>
            </Fragment>
          )
        }}
      </RunContext.Consumer>
    )
  }
}

LogRenderer.propTypes = {
  height: PropTypes.number,
  len: PropTypes.number.isRequired,
  logs: PropTypes.arrayOf(PropTypes.string).isRequired,
  width: PropTypes.number,
}

LogRenderer.defaultProps = {
  height: 0,
  len: 0,
  logs: [],
  width: 0,
}

class LogProcessor extends Component {
  static HACKY_CHAR_TO_PIXEL_RATIO = 37 / 300

  getMaxLineLength = () =>
    round(this.props.width * LogProcessor.HACKY_CHAR_TO_PIXEL_RATIO)

  processLogs = () => {
    const { logs } = this.props

    if (isEmpty(logs)) return []

    const maxLineLength = this.getMaxLineLength()

    return logs.reduce((acc, chunk) => {
      // Split the chunk string by newline chars.
      const split = chunk.getChunk().split("\n")

      // Loop through each split part of the chunk. For each part, if the
      // length of the string is greater than the maxLineLength variable, split
      // the part so each sub-part is less than maxLineLength. Otherwise, push
      // the part to the array to be returned.
      for (let i = 0; i < split.length; i++) {
        const str = split[i]

        if (str.length > maxLineLength) {
          for (let j = 0; j < str.length; j += maxLineLength) {
            acc.push(str.slice(j, j + maxLineLength))
          }
        } else {
          acc.push(str)
        }
      }

      return acc
    }, [])
  }

  areDimensionsValid = () => {
    const { width, height } = this.props

    if (
      width === 0 ||
      width === undefined ||
      height === 0 ||
      height === undefined
    ) {
      return false
    }

    return true
  }

  render() {
    if (this.areDimensionsValid()) {
      const logs = this.processLogs()
      return <LogRenderer {...this.props} logs={logs} len={logs.length} />
    }

    return <span />
  }
}

LogProcessor.propTypes = {
  height: PropTypes.number,
  logs: PropTypes.arrayOf(PropTypes.any).isRequired,
  width: PropTypes.number,
}

LogProcessor.defaultProps = {
  height: window.innerHeight - NAVIGATION_HEIGHT_PX,
  logs: [],
  width: 0,
}

export default props => (
  <ReactResizeDetector
    handleHeight
    handleWidth
    refreshMode="throttle"
    refreshRate={500}
  >
    {(w, h) => <LogProcessor {...props} width={w} height={h} />}
  </ReactResizeDetector>
)
