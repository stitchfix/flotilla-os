import * as React from "react"
import Ansi from "ansi-to-react"
import { Spinner, Pre, Classes, Tag, Checkbox, Icon } from "@blueprintjs/core"
import { LogChunk } from "../types"

type Props = {
  logs: LogChunk[]
  hasRunFinished: boolean
  isLoading: boolean
  height: number
  shouldAutoscroll: boolean
}

class LogRenderer extends React.Component<Props> {
  private CONTAINER_DIV = React.createRef<HTMLDivElement>()

  componentDidMount() {
    this.scrollToBottom()
  }

  componentDidUpdate(prevProps: Props) {
    if (this.shouldScrollToBottom(prevProps, this.props)) {
      this.scrollToBottom()
    }
  }

  scrollToTop = (): void => {
    const container = this.CONTAINER_DIV.current

    if (container) {
      container.scrollTop = 0
    }
  }

  scrollToBottom = (): void => {
    const container = this.CONTAINER_DIV.current

    if (container) {
      container.scrollTop = container.scrollHeight
    }
  }

  shouldScrollToBottom(prev: Props, next: Props) {
    // Handle manual override.
    if (next.shouldAutoscroll === false) return false

    // Handle CloudWatchLogs.
    if (prev.logs.length !== next.logs.length) return true

    // Handle S3 logs (there will only be one chunk).
    if (
      prev.logs.length === 1 &&
      next.logs.length === 1 &&
      prev.logs[0].chunk.length !== next.logs[0].chunk.length
    )
      return true
    return false
  }

  render() {
    const { logs, height, hasRunFinished, isLoading } = this.props

    let loader = <Tag>END OF LOGS</Tag>

    if (!hasRunFinished || isLoading) {
      loader = <Spinner size={Spinner.SIZE_SMALL} />
    }

    return (
      <div
        ref={this.CONTAINER_DIV}
        className="flotilla-logs-container"
        style={{ height }}
      >
        <Pre className={`flotilla-pre ${Classes.DARK}`}>
          {logs.map((l: LogChunk) => (
            // Note: using a CSS classname here as the Ansi component does not
            // support a `style` prop.
            <Ansi linkify={false} key={l.lastSeen} className="flotilla-ansi">
              {l.chunk}
            </Ansi>
          ))}
          <span
            style={{
              display: "flex",
              flexFlow: "row nowrap",
              justifyContent: "flex-start",
              alignItems: "center",
              width: "100%",
              padding: 24,
            }}
          >
            {loader}
          </span>
        </Pre>
      </div>
    )
  }
}

export default LogRenderer
