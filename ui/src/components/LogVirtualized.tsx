import * as React from "react"
import { connect, ConnectedProps as ReduxConnectedProps } from "react-redux"
import { get, isEmpty } from "lodash"
import { FixedSizeList as List } from "react-window"
import Query, { ChildProps } from "./QueryParams"
import LogRow from "./LogVirtualizedRow"
import { LOG_SEARCH_QUERY_KEY } from "../constants"
import LogVirtualizedSearch from "./LogVirtualizedSearch"
import { RootState } from "../state/store"
import { clearSearchState, setMatches } from "../state/search"

type Props = ConnectedProps & ChildProps

/** Renders the processed logs using react-window for performance. */
class LogVirtualized extends React.Component<Props> {
  static defaultProps: Partial<Props> = {
    height: 0,
    logs: [],
    width: 0,
    shouldAutoscroll: true,
  }
  private reactWindowRef = React.createRef<List>()

  componentDidMount() {
    const listRef = this.reactWindowRef.current

    // Scroll to the most recent log.
    if (this.props.shouldAutoscroll === true && listRef) {
      listRef.scrollToItem(this.props.logs.length)
    }
  }

  componentDidUpdate(prevProps: Props) {
    if (
      this.props.shouldAutoscroll === true &&
      prevProps.logs.length !== this.props.logs.length
    ) {
      // Scroll to the most recent log if autoscroll is enabled.
      const listRef = this.reactWindowRef.current
      if (listRef) {
        listRef.scrollToItem(this.props.logs.length)
      }
    }

    const prevSearchQ = get(prevProps.query, LOG_SEARCH_QUERY_KEY)
    const currSearchQ = get(this.props.query, LOG_SEARCH_QUERY_KEY)

    if (prevSearchQ !== currSearchQ) {
      if (isEmpty(currSearchQ)) {
        this.props.dispatch(clearSearchState())
      } else {
        this.search(get(this.props.query, LOG_SEARCH_QUERY_KEY))
      }
    }

    if (prevProps.cursor !== this.props.cursor) {
      this.handleCursorChange()
    }
  }

  handleScrollToTopClick = (): void => {
    const listRef = this.reactWindowRef.current
    if (listRef) {
      listRef.scrollToItem(0)
    }
  }

  handleScrollToBottomClick = (): void => {
    const listRef = this.reactWindowRef.current
    if (listRef) {
      listRef.scrollToItem(this.props.logs.length)
    }
  }

  search(q: string): void {
    const { logs } = this.props

    let matches = []

    for (let i = 0; i < logs.length; i++) {
      const line: string = logs[i]
      const firstIndex = line.indexOf(q)
      // todo: search mroe than first index
      if (firstIndex > -1) {
        const m: [number, number] = [i, firstIndex]
        matches.push(m)
      }
    }

    this.props.dispatch(setMatches({ matches }))
  }

  handleCursorChange(): void {
    const listRef = this.reactWindowRef.current
    if (listRef) {
      const { matches, cursor } = this.props

      if (matches !== null) {
        listRef.scrollToItem(matches[cursor][0])
      }
    }
  }

  render() {
    const { width, height, logs } = this.props

    return (
      <>
        <LogVirtualizedSearch />
        <div className="flotilla-logs-container">
          <List
            ref={this.reactWindowRef}
            height={height}
            itemCount={logs.length}
            itemData={logs}
            itemSize={24}
            width={width}
            overscanCount={100}
          >
            {LogRow}
          </List>
        </div>
      </>
    )
  }
}

const reduxConnector = connect((s: RootState) => s.search)

type ConnectedProps = {
  width: number
  height: number
  logs: string[]
  shouldAutoscroll: boolean
} & ReduxConnectedProps<typeof reduxConnector>

const Connected: React.FC<ConnectedProps> = p => (
  <Query>{q => <LogVirtualized {...p} {...q} />}</Query>
)

export default reduxConnector(Connected)
