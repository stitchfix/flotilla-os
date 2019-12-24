import * as React from "react"
import Ansi from "ansi-to-react"
import { get } from "lodash"
import { ListChildComponentProps } from "react-window"
import { Pre, Classes } from "@blueprintjs/core"
import { useSelector } from "react-redux"
import { RootState } from "../state/store"

const LogVirtualizedRow: React.FC<ListChildComponentProps> = props => {
  const { index, style } = props
  const { matches, cursor } = useSelector((s: RootState) => s.search)

  const [cursorLineNum, cursorCharNum] = get(matches, cursor, [
    undefined,
    undefined,
  ])

  return (
    <Pre
      className={`flotilla-pre ${Classes.DARK}`}
      style={{ ...style, color: cursorLineNum === index ? "pink" : "" }}
    >
      <Ansi className="flotilla-ansi" linkify={false}>
        {get(props, "data", [])[index]}
      </Ansi>
    </Pre>
  )
}

export default LogVirtualizedRow
