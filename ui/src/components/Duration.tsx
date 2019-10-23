import * as React from "react"
import prettyMS from "pretty-ms"
import calculateDuration from "../helpers/calculateDuration"

type Props = {
  start: string
  end: string | undefined | null
}

type State = {
  duration: number
}

class Duration extends React.Component<Props, State> {
  private intervalID: number | undefined
  constructor(props: Props) {
    super(props)
    this.process = this.process.bind(this)
  }

  state = {
    duration: 0,
  }

  componentDidMount() {
    // Immediately process duration on mount.
    this.process()

    // If the end date is undefined, begin interval to process duration.
    if (!this.props.end) {
      this.intervalID = window.setInterval(this.process.bind(this), 1000)
    }
  }

  componentWillUnmount() {
    window.clearInterval(this.intervalID)
  }

  process() {
    const { start, end } = this.props
    this.setState({ duration: calculateDuration(start, end) })
  }

  render() {
    return prettyMS(this.state.duration)
  }
}

export default Duration
