import React, { Component } from "react"
import PropTypes from "prop-types"
import ReactResizeDetector from "react-resize-detector"
import { isEmpty, round } from "lodash"
import { NAVIGATION_HEIGHT_PX } from "../../helpers/styles"
import LogRenderer from "./LogRenderer"

/**
 * The intermediate component between LogRequester and LogRenderer.
 */
class LogProcessor extends Component {
  static HACKY_CHAR_TO_PIXEL_RATIO = 37 / 300

  /**
   * Returns the max number of characters allowed per line.
   */
  getMaxLineLength = () =>
    round(this.props.width * LogProcessor.HACKY_CHAR_TO_PIXEL_RATIO)

  /**
   * Takes the `logs` prop (an array of LogChunk objects), splits each
   * LogChunk's log string according to the available width, and flattens it to
   * an array of strings, which it then passes to LogRenderer to render.
   */
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

  /**
   * Checks whether the dimensions have been set by ReactSizeDetector.
   */
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
      // Only process logs if the dimensions are valid.
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
