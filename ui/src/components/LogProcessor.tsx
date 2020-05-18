import * as React from "react"
import { get } from "lodash"
import ReactResizeDetector from "react-resize-detector"
import WebWorker from "../workers/index"
import LogWorker from "../workers/log.worker"
import { CHAR_TO_PX_RATIO } from "../constants"
import LogVirtualized from "./LogVirtualized"
import { Spinner, Callout } from "@blueprintjs/core"

type ConnectedProps = {
  logs: string
  hasRunFinished: boolean
}

type Props = ConnectedProps & {
  width: number
  height: number
}

type State = {
  isProcessing: boolean
  processedLogs: string[]
}

export class LogProcessor extends React.Component<Props, State> {
  private logWorker: any
  constructor(props: Props) {
    super(props)

    // Instantiate worker and add event listener.
    if (process.env.NODE_ENV !== "test") {
      this.logWorker = new WebWorker(LogWorker)
      this.logWorker.addEventListener("message", (evt: any) => {
        this.setState({
          processedLogs: get(evt, "data", []),
          isProcessing: false,
        })
      })
    }
  }

  state: State = {
    isProcessing: false,
    processedLogs: [],
  }

  componentDidMount() {
    this.processLogs()
  }

  componentDidUpdate(prevProps: Props) {
    // If the log length or container width change, re-process logs. Note: the
    // container height has no effect on this.
    if (
      prevProps.logs.length !== this.props.logs.length ||
      prevProps.width !== this.props.width
    ) {
      this.processLogs()
    }
  }

  /** Returns the max number of characters allowed per line. */
  getMaxLineLength = (): number =>
    Math.floor(this.props.width * CHAR_TO_PX_RATIO)

  /** Send props.logs to web worker for processing. */
  processLogs(): void {
    const { logs } = this.props

    // Early exit if running tests or no logs.
    if (process.env.NODE_ENV === "test" || logs.length === 0) return

    this.setState({ isProcessing: true })
    this.logWorker.postMessage({
      logs,
      maxLen: this.getMaxLineLength(),
    })
  }

  render() {
    const { width, height, hasRunFinished } = this.props
    let { isProcessing, processedLogs } = this.state

    processedLogs = processedLogs.map((el) => el + "\n")

    // If no existing logs and processing, return spinner.
    if (isProcessing && processedLogs.length === 0) {
      return (
        <Callout>
          <div style={{ display: "flex" }}>
            Optimizing... <Spinner size={Spinner.SIZE_SMALL} />
          </div>
        </Callout>
      )
    }

    return (
      <LogVirtualized
        logs={processedLogs}
        width={width}
        height={height}
        hasRunFinished={hasRunFinished}
      />
    )
  }
}

const Connected: React.FC<ConnectedProps> = props => (
  <ReactResizeDetector
    handleHeight
    handleWidth
    refreshMode="throttle"
    refreshRate={1000}
  >
    {({ width }: { width?: number; height?: number }) => (
      <LogProcessor
        logs={props.logs}
        hasRunFinished={props.hasRunFinished}
        width={width || 500}
        height={600}
      />
    )}
  </ReactResizeDetector>
)

export default Connected
