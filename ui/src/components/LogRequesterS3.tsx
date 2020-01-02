import * as React from "react"
import { connect, ConnectedProps } from "react-redux"
import api from "../api"
import LogProcessor from "./LogProcessor"
import { RunStatus } from "../types"
import { LOG_FETCH_INTERVAL_MS } from "../constants"
import ErrorCallout from "./ErrorCallout"
import { RootState } from "../state/store"
import { setHasLogs } from "../state/runView"

const connected = connect((state: RootState) => state.runView)

type Props = {
  status: RunStatus | undefined
  runID: string
} & ConnectedProps<typeof connected>

type State = {
  logs: string
  isLoading: boolean
  error: any
}

const initialState: State = {
  logs: "",
  isLoading: false,
  error: false,
}

class LogRequesterS3 extends React.PureComponent<Props, State> {
  private requestInterval: number | undefined
  state = initialState

  componentDidMount() {
    this.initialize()
  }

  componentDidUpdate(prevProps: Props) {
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
  }

  componentWillUnmount() {
    window.clearInterval(this.requestInterval)
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
    const { runID, hasLogs } = this.props

    this.setState({ isLoading: true })

    api
      .getRunLogRaw({ runID })
      .then((logs: string) => {
        this.setState({
          isLoading: false,
          error: false,
          logs,
        })

        if (hasLogs === false && logs.length > 0) {
          this.props.dispatch(setHasLogs())
        }
      })
      .catch(error => {
        this.clearRequestInterval()
        this.setState({ isLoading: false, error })
      })
  }

  render() {
    const { status } = this.props
    const { error, logs } = this.state
    if (error) return <ErrorCallout error={error} />
    return (
      <LogProcessor logs={logs} hasRunFinished={status === RunStatus.STOPPED} />
    )
  }
}

export default connected(LogRequesterS3)
