import * as React from "react"
import Ansi from "ansi-to-react"
import { get } from "lodash"
import { ListChildComponentProps } from "react-window"
import { Pre, Classes, Colors, Tag, Spinner } from "@blueprintjs/core"

const LogVirtualizedRow: React.FC<ListChildComponentProps> = props => {
  const { index, style, data } = props
  const lines: string[] = get(data, "lines", [])
  const hasRunFinished: boolean = get(data, "hasRunFinished", false)
  const searchMatches: [number, number][] = get(data, "searchMatches", [])
  const searchCursor: number = get(data, "searchCursor", 0)
  const searchCursorLineNumber = get(searchMatches, [searchCursor, 0], null)

  // Note: the last item will be a spinner or a tag indicating the end of logs.
  if (index === lines.length) {
    if (hasRunFinished) {
      return (
        <div style={style}>
          <Tag>END OF LOGS</Tag>
        </div>
      )
    }

    return (
      <div style={style}>
        <Spinner size={Spinner.SIZE_SMALL} />
      </div>
    )
  }

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
