import * as React from "react"
import getRunDuration from "../../helpers/getRunDuration"

interface IRunDurationCounterProps {
  started_at?: string
  finished_at?: string
}

interface IRunDurationCounterState {
  duration: number | string
}

class RunDurationCounter extends React.PureComponent<
  IRunDurationCounterProps,
  IRunDurationCounterState
> {
  private interval = -1
  private DURATION_INIT_STATE = "-"
  state = {
    duration: this.DURATION_INIT_STATE,
  }
  componentDidMount() {
    this.startInterval()
  }

  componentDidUpdate(prevProps: IRunDurationCounterProps) {
    if (
      !prevProps.started_at &&
      this.props.started_at &&
      this.interval === -1
    ) {
      this.startInterval()
      return
    }

    if (!prevProps.finished_at && this.props.finished_at) {
      this.stopInterval()
      return
    }
  }

  componentWillUnmount() {
    this.stopInterval()
  }

  startInterval = (): void => {
    if (this.props.started_at) {
      this.interval = window.setInterval(() => {
        this.setState({
          duration: getRunDuration({
            started_at: this.props.started_at,
            finished_at: this.props.finished_at,
          }),
        })
      }, 1000)
    }
  }

  stopInterval = (): void => {
    if (this.interval !== -1) {
      window.clearInterval(this.interval)
    }
  }

  render() {
    return this.state.duration
  }
}

export default RunDurationCounter
