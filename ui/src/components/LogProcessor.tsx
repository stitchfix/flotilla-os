import * as React from "react"
import ReactResizeDetector from "react-resize-detector"
import { get, round, isEmpty } from "lodash"
import LogRendererOptimized from "./LogRendererOptimized"
import { LogChunk } from "../types"
import WebWorker from "../workers/index"
import LogWorker from "../workers/log.worker"
import { Button, Spinner } from "@blueprintjs/core"
import { DebounceInput } from "react-debounce-input"
import QueryParams, { ChildProps } from "./QueryParams"
import { LOG_SEARCH_QUERY_KEY } from "../constants"

type ConnectedProps = {
  logs: LogChunk[]
}

type Props = ConnectedProps &
  ChildProps & {
    width: number
    height: number
  }

type State = {
  logs: string[]
  isSearching: boolean
  matches: Array<[number, number]>
}

/**
 * The intermediate component between LogRequester and LogRendererOptimized.
 * This component is responsible for slicing the logs into smaller pieces, each
 * of which will be rendered into a LowRow component.
 */
class LogProcessor extends React.PureComponent<Props, State> {
  static HACKY_CHAR_TO_PIXEL_RATIO = 40 / 300
  private logWorker: any

  constructor(props: Props) {
    super(props)
    this.logWorker = new WebWorker(LogWorker)
    this.logWorker.addEventListener("message", (evt: any) => {
      console.log("received message from worker")
      this.setState({ logs: evt.data })
    })
  }

  state = {
    logs: [],
    isSearching: false,
    matches: [],
  }

  componentDidMount() {
    this.processLogs()
  }

  componentDidUpdate(prevProps: Props) {
    if (prevProps.logs.length !== this.props.logs.length) {
      this.processLogs()
    }

    // if (
    //   prevProps.width !== this.props.width ||
    //   prevProps.height !== this.props.height
    // ) {
    //   this.processLogs()
    // }

    const prevSearchQ = get(prevProps.query, LOG_SEARCH_QUERY_KEY)
    const currSearchQ = get(this.props.query, LOG_SEARCH_QUERY_KEY)

    if (prevSearchQ !== currSearchQ) {
      if (isEmpty(currSearchQ)) {
        this.setState({ matches: [] })
      } else {
        this.search(get(this.props.query, LOG_SEARCH_QUERY_KEY))
      }
    }
  }

  /**
   * Returns the max number of characters allowed per line.
   */
  getMaxLineLength = (): number =>
    round(this.props.width * LogProcessor.HACKY_CHAR_TO_PIXEL_RATIO)

  /**
   * Takes the `logs` prop (an array of LogChunk objects), splits each
   * LogChunk's log string according to the available width, and flattens it to
   * an array of strings, which it then passes to LogRendererOptimized to render.
   */
  processLogs = (): void => {
    const { logs } = this.props
    console.log("sending preprocessed logs to worker")
    this.logWorker.postMessage({
      chunks: logs,
      maxLen: this.getMaxLineLength(),
    })
  }

  search(q: string): void {
    const { logs } = this.state

    this.setState({ isSearching: true, matches: [] }, () => {
      let matches = []

      for (let i = 0; i < logs.length; i++) {
        const line: string = logs[i]
        const firstIndex = line.indexOf(q)
        if (firstIndex > -1) {
          const m: [number, number] = [i, firstIndex]
          matches.push(m)
        }
      }

      this.setState({ isSearching: false, matches })
    })
  }

  render() {
    const { query, setQuery, width, height } = this.props
    const { logs, isSearching, matches } = this.state

    return (
      <>
        <div>
          <DebounceInput
            value={get(query, LOG_SEARCH_QUERY_KEY, "")}
            onChange={evt => {
              setQuery({ ...query, [LOG_SEARCH_QUERY_KEY]: evt.target.value })
            }}
            debounceTimeout={500}
            className="bp3-input"
          />
          {isSearching === true && <Spinner size={Spinner.SIZE_LARGE} />}
          <div>
            <div>number of matches: {matches.length}</div>
          </div>
        </div>
        <LogRendererOptimized
          logs={logs}
          len={logs.length}
          width={width}
          height={height}
        />
      </>
    )
  }
}

const Connected: React.FC<ConnectedProps> = props => (
  <ReactResizeDetector
    handleHeight
    handleWidth
    refreshMode="throttle"
    refreshRate={500}
  >
    {({ width, height }: { width: number; height: number }) => {
      console.log(width, height)
      return (
        <QueryParams>
          {({ query, setQuery }) => (
            <LogProcessor
              logs={props.logs}
              width={width || 500}
              height={height || 500}
              query={query}
              setQuery={setQuery}
            />
          )}
        </QueryParams>
      )
    }}
  </ReactResizeDetector>
)

export default Connected
