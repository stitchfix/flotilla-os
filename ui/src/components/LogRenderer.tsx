import * as React from "react"
import Ansi from "ansi-to-react"
import { Spinner, Pre, Classes, Tag } from "@blueprintjs/core"
import { LogChunk } from "../types"

type Props = {
  logs: LogChunk[]
  hasRunFinished: boolean
  isLoading: boolean
  height: number
}

class LogRenderer extends React.Component<Props> {
  private CONTAINER_DIV = React.createRef<HTMLDivElement>()
  state = {
    shouldAutoscroll: true,
  }

  componentDidMount() {
    this.scrollToBottom()
  }

  componentDidUpdate(prevProps: Props) {
    if (prevProps.logs.length !== this.props.logs.length) {
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

  // shouldRenderLoader(): boolean {
  //   const { hasRunFinished, isLoading } = this.props
  //   if (hasRunFinished === false || isLoading === true) return true
  //   return false
  // }

  render() {
    const { logs, height, hasRunFinished } = this.props

    let loader = <Tag>END OF LOGS</Tag>

    if (!hasRunFinished) {
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
