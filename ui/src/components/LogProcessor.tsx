import * as React from "react"
import ReactResizeDetector from "react-resize-detector"
import { round } from "lodash"
import LogRendererOptimized from "./LogRendererOptimized"
import { LogChunk } from "../types"
import WebWorker from "../workers/index"
import LogWorker from "../workers/log.worker"

type ConnectedProps = {
  logs: LogChunk[]
}

type Props = ConnectedProps & {
  width: number
  height: number
}

type State = {
  logs: string[]
}

/**
 * The intermediate component between LogRequester and LogRendererOptimized.
 * This component is responsible for slicing the logs into smaller pieces, each
 * of which will be rendered into a LowRow component.
 */
class LogProcessor extends React.PureComponent<Props, State> {
  static HACKY_CHAR_TO_PIXEL_RATIO = 37 / 300
  private logWorker: any

  constructor(props: Props) {
    super(props)
    this.logWorker = new WebWorker(LogWorker)
  }

  state = {
    logs: [],
  }

  componentDidMount() {
    this.processLogs()
  }

  componentDidUpdate(prevProps: Props) {
    if (prevProps.logs.length !== this.props.logs.length) {
      this.processLogs()
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
    this.logWorker.addEventListener("message", (evt: any) => {
      console.log("received message from worker")
      this.setState({ logs: evt.data })
    })
  }

  render() {
    const { logs } = this.state
    return <LogRendererOptimized logs={logs} len={logs.length} />
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
      return <LogProcessor logs={props.logs} width={500} height={500} />
    }}
  </ReactResizeDetector>
)

export default Connected
