import * as React from "react"
import Ansi from "ansi-to-react"
import { get } from "lodash"
import { Pre } from "../styled/Monospace"
import Loader from "../styled/Loader"
import RunContext from "../Run/RunContext"
import { flotillaRunStatuses, flotillaUIIntents } from "../../types"
import { ListChildComponentProps } from "react-window"

/**
 * Renders a line of logs. Will also render a spinner as the last child if
 * the run is still active.
 */
class LogRow extends React.PureComponent<ListChildComponentProps> {
  render() {
    const { index, style } = this.props

    return (
      <RunContext.Consumer>
        {ctx => {
          const isStopped =
            get(ctx, ["data", "status"]) === flotillaRunStatuses.STOPPED

          if (!isStopped && index === get(this.props, "data", []).length - 1) {
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
                <Loader intent={flotillaUIIntents.PRIMARY} />
              </span>
            )
          }

          return (
            <Pre style={style}>
              <Ansi className="flotilla-ansi">
                {get(this.props, "data", [])[index]}
              </Ansi>
            </Pre>
          )
        }}
      </RunContext.Consumer>
    )
  }
}

export default LogRow
