import * as React from "react"
import Ansi from "ansi-to-react"
import { get } from "lodash"
import { ListChildComponentProps } from "react-window"
import { Pre, Classes } from "@blueprintjs/core"

/**
 * Renders a line of logs. Will also render a spinner as the last child if
 * the run is still active.
 */
const LogRow: React.FC<ListChildComponentProps> = props => {
  const { index, style } = props
  return (
    <Pre className={`flotilla-pre ${Classes.DARK}`} style={style}>
      <Ansi className="flotilla-ansi" linkify={false}>
        {get(props, "data", [])[index]}
      </Ansi>
    </Pre>
  )
}

export default LogRow
