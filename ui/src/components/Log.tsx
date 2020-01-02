import * as React from "react"
import { connect, ConnectedProps } from "react-redux"
import Ansi from "ansi-to-react"
import { Spinner, Pre, Classes, Tag } from "@blueprintjs/core"
import { RootState } from "../state/store"

const connector = connect((state: RootState) => state.runView)

type Props = {
  logs: string
  hasRunFinished: boolean
  isLoading: boolean
} & ConnectedProps<typeof connector>

class Log extends React.Component<Props> {
  private CONTAINER_DIV = React.createRef<HTMLDivElement>()

  componentDidMount() {
    if (this.props.shouldAutoscroll) {
      this.scrollToBottom()
    }
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
    const { logs, hasRunFinished, isLoading } = this.props

    let loader = <Tag>END OF LOGS</Tag>

    if (!hasRunFinished || isLoading) {
      loader = <Spinner size={Spinner.SIZE_SMALL} />
    }

    return (
      <div ref={this.CONTAINER_DIV} className="flotilla-logs-container">
        <Pre className={`flotilla-pre ${Classes.DARK}`}>
          <Ansi linkify={false} className="flotilla-ansi">
            {logs}
          </Ansi>
        </Pre>
        <div className="flotilla-logs-loader-container">{loader}</div>
      </div>
    )
  }
}

export default connector(Log)
