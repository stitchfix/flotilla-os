import React, { Component } from "react"
import PropTypes from "prop-types"
import { FixedSizeList as List } from "react-window"
import ReactResizeDetector from "react-resize-detector"
import { isEmpty, round } from "lodash"
import LogRow from "./LogRow"
import { TOPBAR_HEIGHT_PX } from "../../constants/styles"

class LogRenderer extends Component {
  static HACKY_CHAR_TO_PIXEL_RATIO = 37 / 300

  getMaxLineLength = () =>
    round(this.props.width * LogRenderer.HACKY_CHAR_TO_PIXEL_RATIO)

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

  render() {
    console.log(window.innerWidth, this.props.width)
    console.log(window.innerHeight, this.props.height)
    if (this.areDimensionsValid()) {
      const { width, height } = this.props
      const logs = this.processLogs()
      const len = logs.length

      return (
        <List
          height={height}
          itemCount={len}
          itemData={logs}
          itemSize={20}
          width={width}
        >
          {LogRow}
        </List>
      )
    }

    return <span />
  }
}

LogRenderer.propTypes = {
  height: PropTypes.number,
  logs: PropTypes.arrayOf(PropTypes.any).isRequired,
  width: PropTypes.number,
}

LogRenderer.defaultProps = {
  height: 0,
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
    {(w, h) => <LogRenderer {...props} width={w} height={h} />}
  </ReactResizeDetector>
)
