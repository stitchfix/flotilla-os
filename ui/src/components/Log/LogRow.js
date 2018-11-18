import React, { PureComponent } from "react"
import PropTypes from "prop-types"
import Ansi from "ansi-to-react"
import { get } from "lodash"
import Pre from "../styled/Pre"
import Loader from "../styled/Loader"
import runStatusTypes from "../../constants/runStatusTypes"
import RunContext from "../Run/RunContext"

/**
 * Renders a line of logs. Will also render a spinner as the last child if
 * the run is still active.
 */
class LogRow extends PureComponent {
  render() {
    const { data, index, style } = this.props

    return (
      <RunContext.Consumer>
        {ctx => {
          const isStopped =
            get(ctx, ["data", "status"]) === runStatusTypes.stopped

          if (!isStopped && index === data.length - 1) {
            return (
              <span
                style={{
                  ...style,
                  display: "flex",
                  flexFlow: "row nowrap",
                  justifyContent: "flex-start",
                  alignItems: "center",
                  width: "100%",
                }}
              >
                <Loader />
              </span>
            )
          }

          return (
            <Pre style={style}>
              <Ansi>{data[index]}</Ansi>
            </Pre>
          )
        }}
      </RunContext.Consumer>
    )
  }
}

LogRow.propTypes = {
  data: PropTypes.array,
  index: PropTypes.number,
  style: PropTypes.object,
}

export default LogRow
