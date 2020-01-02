import * as React from "react"
import Ansi from "ansi-to-react"
import { get } from "lodash"
import { ListChildComponentProps } from "react-window"
import { Pre, Classes, Colors } from "@blueprintjs/core"

const LogVirtualizedRow: React.FC<ListChildComponentProps> = props => {
  const { index, style, data } = props
  const lines: string[] = get(data, "lines", [])
  const searchMatches: [number, number][] = get(data, "searchMatches", [])
  const searchCursor: number = get(data, "searchCursor", 0)
  const searchCursorLineNumber = get(searchMatches, [searchCursor, 0], null)

  return (
    <Pre
      className={`flotilla-pre ${Classes.DARK}`}
      style={{
        ...style,
        color: searchCursorLineNumber === index ? Colors.GOLD5 : "",
      }}
    >
      <Ansi className="flotilla-ansi" linkify={false}>
        {lines[index]}
      </Ansi>
    </Pre>
  )
}

export default LogVirtualizedRow
