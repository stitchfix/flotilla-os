import * as React from "react"
import ReactResizeDetector from "react-resize-detector"
import WebWorker from "../workers/index"
import LogWorker from "../workers/log.worker"
import { CHAR_TO_PX_RATIO } from "../constants"
import LogVirtualized from "./LogVirtualized"

type ConnectedProps = {
  logs: string
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
    this.logWorker = new WebWorker(LogWorker)
    this.logWorker.addEventListener("message", (evt: any) => {
      this.setState({ isProcessing: false, processedLogs: evt.data })
    })
  }

  state = {
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

    this.setState({ isProcessing: true }, () => {
      this.logWorker.postMessage({
        logs,
        maxLen: this.getMaxLineLength(),
      })
    })
  }

  render() {
    const { width, height } = this.props
    const { processedLogs } = this.state

    return (
      <LogVirtualized
        logs={processedLogs}
        width={width}
        height={height}
        shouldAutoscroll
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
    {({ width, height }: { width?: number; height?: number }) => (
      <LogProcessor logs={props.logs} width={width || 500} height={800} />
    )}
  </ReactResizeDetector>
)

export default Connected
