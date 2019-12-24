import * as React from "react"
import api from "../api"
import LogProcessor from "./LogProcessor"
import { RunStatus } from "../types"
import { LOG_FETCH_INTERVAL_MS } from "../constants"
import ErrorCallout from "./ErrorCallout"

type Props = {
  status: RunStatus | undefined
  runID: string
  height: number
  setHasLogs: () => void
  shouldAutoscroll: boolean
}

type State = {
  logs: string
  isLoading: boolean
  error: any
  hasLogs: boolean
}

const initialState: State = {
  logs: "",
  isLoading: false,
  error: false,
  hasLogs: false,
}

class LogRequesterS3 extends React.PureComponent<Props, State> {
  private requestInterval: number | undefined
  state = initialState

  componentDidMount() {
    this.initialize()
  }

  componentDidUpdate(prevProps: Props, prevState: State) {
    if (prevProps.runID !== this.props.runID) {
      this.handleRunIDChange()
      return
    }

    if (
      prevProps.status !== RunStatus.STOPPED &&
      this.props.status === RunStatus.STOPPED
    ) {
      this.clearRequestInterval()
    }

    if (prevState.hasLogs === false && this.state.hasLogs === true) {
      this.props.setHasLogs()
    }
  }

  componentWillUnmount() {
    this.clearRequestInterval()
  }

  setRequestInterval = (): void => {
    this.requestInterval = window.setInterval(
      this.requestLogs,
      LOG_FETCH_INTERVAL_MS
    )
  }

  clearRequestInterval = () => {
    window.clearInterval(this.requestInterval)
  }

  initialize() {
    this.requestLogs()

    if (this.props.status !== RunStatus.STOPPED) {
      this.setRequestInterval()
    }
  }

  handleRunIDChange() {
    // Clear request interval
    this.clearRequestInterval()

    // Reset state.
    this.setState(initialState, () => {
      // Initialize, as if the component just mounted.
      this.initialize()
    })
  }

  requestLogs = () => {
    const { runID } = this.props

    this.setState({ isLoading: true })

    api
      .getRunLogRaw({ runID })
      .then((logs: string) => {
        this.setState({
          isLoading: false,
          error: false,
          logs,
          hasLogs: logs.length > 0,
        })
      })
      .catch(error => {
        this.clearRequestInterval()
        this.setState({ isLoading: false, error })
      })
  }

  render() {
    const { error, logs } = this.state

    if (error) return <ErrorCallout error={error} />

    return <LogProcessor logs={logs} />
  }
}

export default LogRequesterS3
