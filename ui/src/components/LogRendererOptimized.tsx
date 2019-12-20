import * as React from "react"
import { FixedSizeList as List } from "react-window"
import LogRow from "./LogRow"

type Props = {
  len: number
  width: number
  height: number
  logs: string[]
}

type State = {
  shouldAutoscroll: boolean
}

/** Renders the processed logs using react-window for performance. */
class LogRendererOptimized extends React.Component<Props, State> {
  static defaultProps: Partial<Props> = {
    height: 0,
    len: 0,
    logs: [],
    width: 0,
  }
  private LIST_REF = React.createRef<any>()
  state = {
    shouldAutoscroll: true,
  }

  componentDidMount() {
    const listRef = this.LIST_REF.current

    // Scroll to the most recent log.
    if (listRef) {
      listRef.scrollToItem(this.props.len)
    }
  }

  componentDidUpdate(prevProps: Props) {
    if (
      this.state.shouldAutoscroll === true &&
      prevProps.len !== this.props.len
    ) {
      // Scroll to the most recent log if autoscroll is enabled.
      const listRef = this.LIST_REF.current
      if (listRef) {
        listRef.scrollToItem(this.props.len)
      }
    }
  }

  handleScrollToTopClick = (): void => {
    const listRef = this.LIST_REF.current
    if (listRef) {
      listRef.scrollToItem(0)
    }
  }

  handleScrollToBottomClick = (): void => {
    const listRef = this.LIST_REF.current
    if (listRef) {
      listRef.scrollToItem(this.props.len)
    }
  }

  render() {
    const { width, height, logs, len } = this.props

    return (
      <div className="flotilla-logs-container">
        <List
          ref={this.LIST_REF}
          height={800}
          itemCount={logs.length}
          itemData={logs}
          itemSize={24}
          width={width}
          overscanCount={100}
          // style={{ marginTop: RUN_BAR_HEIGHT_PX }}
        >
          {LogRow}
        </List>
      </div>
    )
  }
}

export default LogRendererOptimized
