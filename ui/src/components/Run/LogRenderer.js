import React, { Component } from "react"
import PropTypes from "prop-types"
import { FixedSizeList as List } from "react-window"
import { isEmpty, round } from "lodash"
import LogChunk from "./LogChunk"
import LogRow from "./LogRow"

class LogRenderer extends Component {
  static HACKY_CHAR_TO_PIXEL_RATIO = 37 / 300

  state = {
    width: 1000,
    height: 500,
  }

  componentDidMount() {}

  getMaxLineLength = () =>
    round(this.state.width * LogRenderer.HACKY_CHAR_TO_PIXEL_RATIO)

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
    const { width, height } = this.state
    const logs = this.processLogs()
    const len = logs.length

    return (
      <div style={{ marginLeft: 50 }}>
        <List
          height={height}
          itemCount={len}
          itemSize={20}
          width={width}
          itemData={logs}
        >
          {LogRow}
        </List>
      </div>
    )
  }
}

LogRenderer.propTypes = {
  logs: PropTypes.arrayOf(PropTypes.any).isRequired,
}

LogRenderer.defaultProps = {
  logs: [],
}

export default LogRenderer
