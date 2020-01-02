import * as React from "react"
import { FixedSizeList as List } from "react-window"
import { connect, ConnectedProps } from "react-redux"
import { get } from "lodash"
import LogRow from "./LogVirtualizedRow"
import LogVirtualizedSearch from "./LogVirtualizedSearch"
import { RootState } from "../state/store"

const connected = connect((state: RootState) => state.runView)

export type Props = {
  width: number
  height: number
  logs: string[]
  hasRunFinished: boolean
} & ConnectedProps<typeof connected>

type State = {
  isSearchInputFocused: boolean
  searchMatches: [number, number][] // [line number, char index]
  searchCursor: number
}

enum KeyCode {
  F = 70,
  ESC = 27,
  ENTER = 13,
}

/** Renders the processed logs using react-window for performance. */
export class LogVirtualized extends React.Component<Props, State> {
  static defaultProps: Partial<Props> = {
    height: 0,
    logs: [],
    width: 0,
  }
  private reactWindowRef = React.createRef<List>()
  private searchInputRef = React.createRef<HTMLInputElement>()

  constructor(props: Props) {
    super(props)
    this.search = this.search.bind(this)
    this.handleCursorChange = this.handleCursorChange.bind(this)
    this.handleIncrementCursor = this.handleIncrementCursor.bind(this)
    this.handleDecrementCursor = this.handleDecrementCursor.bind(this)
    this.handleKeydown = this.handleKeydown.bind(this)
  }

  state: State = {
    isSearchInputFocused: false,
    searchMatches: [],
    searchCursor: 0,
  }

  componentDidMount() {
    window.addEventListener("keydown", this.handleKeydown)

    // Scroll to the most recent log.
    if (this.props.shouldAutoscroll === true) {
      this.scrollTo(this.props.logs.length - 1, "end")
    }
  }

  componentDidUpdate(prevProps: Props, prevState: State) {
    if (prevState.searchCursor !== this.state.searchCursor) {
      this.handleCursorChange()
    }

    if (
      this.props.shouldAutoscroll === true &&
      prevProps.logs.length !== this.props.logs.length
    ) {
      this.scrollTo(this.props.logs.length - 1, "end")
    }
  }

  componentWillUnmount() {
    window.removeEventListener("keydown", this.handleKeydown)
  }

  /**
   * Given a valid query (length > 0), this method will iterate through
   * this.props.logs (string[]) and push the index of the first occurence of
   * the query for each line into the `matches` array.
   */
  search(q: string): void {
    let matches = []

    if (q.length > 0) {
      const { logs } = this.props

      for (let i = 0; i < logs.length; i++) {
        const line: string = logs[i]
        const firstIndex = line.indexOf(q)
        // todo: search more than first index.
        if (firstIndex > -1) {
          const m: [number, number] = [i, firstIndex]
          matches.push(m)
        }
      }
    }

    this.setState({ searchMatches: matches, searchCursor: 0 })
  }

  handleCursorChange(): void {
    const { searchMatches, searchCursor } = this.state

    // If search cursor is within bounds, scroll to the item.
    if (searchCursor >= 0 && searchCursor < searchMatches.length) {
      const lineNumber = get(searchMatches, [searchCursor, 0], 0)
      this.scrollTo(lineNumber, "center")
    }
  }

  handleIncrementCursor(): void {
    if (this.state.searchMatches.length > 0) {
      this.setState(prev => ({
        searchCursor:
          prev.searchCursor === this.state.searchMatches.length - 1
            ? 0
            : prev.searchCursor + 1,
      }))
    }
  }

  handleDecrementCursor(): void {
    if (this.state.searchMatches.length > 0) {
      this.setState(prev => ({
        searchCursor:
          prev.searchCursor === 0
            ? this.state.searchMatches.length - 1
            : prev.searchCursor - 1,
      }))
    }
  }

  handleKeydown(evt: KeyboardEvent) {
    const { isSearchInputFocused } = this.state

    // If the search component is visible and the user hits the escape key,
    // reset search state (hide input, reset matches to an empty array, etc.)
    if (evt.keyCode === KeyCode.ESC && isSearchInputFocused) {
      this.resetSearchState()
      return
    }

    // Handle cmd-f.
    if (evt.keyCode === KeyCode.F && evt.metaKey) {
      evt.preventDefault()
      evt.stopPropagation()
      this.searchInputFocus()
      return
    }

    // If search input is focused and the enter key is pressed, jump to the
    // next search match.
    if (evt.keyCode === KeyCode.ENTER && isSearchInputFocused) {
      this.handleIncrementCursor()
      return
    }
  }

  resetSearchState(): void {
    this.setState({
      isSearchInputFocused: false,
      searchMatches: [],
      searchCursor: 0,
    })
  }

  searchInputFocus() {
    if (this.searchInputRef.current) {
      this.searchInputRef.current.focus()
    }
  }

  scrollTo(
    line: number,
    align?: "auto" | "smart" | "center" | "end" | "start" | undefined
  ) {
    const listRef = this.reactWindowRef.current
    if (listRef) {
      listRef.scrollToItem(line, align)
    }
  }

  render() {
    const { width, height, logs, hasRunFinished } = this.props
    const { searchMatches, searchCursor } = this.state

    return (
      <div className="flotilla-logs-virtualized-container">
        <LogVirtualizedSearch
          onChange={this.search}
          onFocus={() => {
            this.setState({ isSearchInputFocused: true })
          }}
          onBlur={() => {
            this.setState({ isSearchInputFocused: false })
          }}
          onIncrement={this.handleIncrementCursor}
          onDecrement={this.handleDecrementCursor}
          inputRef={this.searchInputRef}
          cursorIndex={searchCursor}
          totalMatches={searchMatches.length}
        />
        <div className="flotilla-logs-container">
          <List
            ref={this.reactWindowRef}
            height={height}
            itemCount={logs.length + 1}
            itemData={{
              lines: logs,
              searchMatches,
              searchCursor,
              hasRunFinished,
            }}
            itemSize={24}
            width={width}
            overscanCount={100}
          >
            {LogRow}
          </List>
        </div>
      </div>
    )
  }
}

export default connected(LogVirtualized)
