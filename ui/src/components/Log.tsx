import * as React from "react"
import Ansi from "ansi-to-react"
import {
  Spinner,
  Pre,
  Classes,
  Tag,
  Checkbox,
  Icon,
  Button,
  ButtonGroup,
} from "@blueprintjs/core"

type Props = {
  logs: string
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
    if (prev.logs.length !== next.logs.length) return true
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
        {/* <Pre className={`flotilla-pre ${Classes.DARK}`}>
          <Ansi linkify={false} className="flotilla-ansi">
            {logs}
          </Ansi>
        </Pre>
        <div className="flotilla-logs-loader-container">{loader}</div> */}
      </div>
    )
  }
}

export default LogRenderer
