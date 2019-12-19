import * as React from "react"
import { FixedSizeList as List } from "react-window"
import { get } from "lodash"
import LogRow from "./LogRow"

interface ILogRendererProps {
  len: number
  width: number
  height: number
  logs: string[]
}

interface ILogRendererState {
  shouldAutoscroll: boolean
}

/** Renders the processed logs using react-window for performance. */
class LogRenderer extends React.PureComponent<
  ILogRendererProps,
  ILogRendererState
> {
  static defaultProps: Partial<ILogRendererProps> = {
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

  componentDidUpdate(prevProps: ILogRendererProps) {
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

  toggleShouldAutoscroll = (): void => {
    this.setState(prev => ({ shouldAutoscroll: !prev.shouldAutoscroll }))
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
      <List
        ref={this.LIST_REF}
        height={height}
        itemCount={len}
        itemData={logs}
        itemSize={20}
        width={width}
        overscanCount={100}
        // style={{ marginTop: RUN_BAR_HEIGHT_PX }}
      >
        {LogRow}
      </List>
    )
  }
}

export default LogRenderer
