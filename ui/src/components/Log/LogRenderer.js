import React, { Component, createRef, Fragment } from "react"
import PropTypes from "prop-types"
import { FixedSizeList as List } from "react-window"
import { get } from "lodash"
import LogRow from "./LogRow"
import { RUN_BAR_HEIGHT_PX } from "../../helpers/styles"
import RunBar from "../Run/RunBar"
import RunContext from "../Run/RunContext"
import runStatusTypes from "../../helpers/runStatusTypes"

// Create a ref for the FixedSizeList component.
const LIST_REF = createRef()

/**
 * Renders the processed logs using react-window for performance.
 */
class LogRenderer extends Component {
  state = {
    shouldAutoscroll: true,
  }

  componentDidMount() {
    // Scroll to the most recent log.
    LIST_REF.current.scrollToItem(this.props.len)
  }

  componentDidUpdate(prevProps) {
    if (
      this.state.shouldAutoscroll === true &&
      prevProps.len !== this.props.len
    ) {
      // Scroll to the most recent log if autoscroll is enabled.
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

export default LogRenderer
