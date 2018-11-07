import React, { Component } from "react"
import PropTypes from "prop-types"
import { get, has } from "lodash"
import Card from "./Card"
import EnhancedRunStatus from "./EnhancedRunStatus"
import FormGroup from "./FormGroup"
import runStatusTypes from "../constants/runStatusTypes"
import getRunDuration from "../utils/getRunDuration"

class RunStatusBar extends Component {
  static propTypes = {
    exitCode: PropTypes.any,
    finishedAt: PropTypes.string,
    startedAt: PropTypes.string,
    status: PropTypes.oneOf(Object.values(runStatusTypes)),
  }
  state = { duration: "-" }
  componentDidMount() {
    this.renderedDuration = window.setInterval(
      this.calculateDuration.bind(this),
      1000
    )
  }
  componentDidUpdate(prevProps, prevState) {
    if (!has(prevProps, "finishedAt") && has(this.props, "finishedAt")) {
      this.stopDurationInterval()
    }
  }
  componentWillUnmount() {
    this.stopDurationInterval()
  }
  stopDurationInterval() {
    window.clearInterval(this.renderedDuration)
  }
  calculateDuration() {
    // Return if task hasn't started.
    if (!has(this.props, "startedAt")) {
      return false
    }

    const start = this.props.startedAt
    const end = get(this.props, "finishedAt", new Date())
    const duration = getRunDuration({
      started_at: start,
      finished_at: end,
    })

    this.setState({ duration })
  }
  render() {
    return (
      <Card>
        <div className="flex ff-rn j-sb a-c full-width with-horizontal-child-margin">
          <FormGroup isStatic label="Status">
            <EnhancedRunStatus
              status={get(this.props, "status", "")}
              exitCode={get(this.props, "exitCode", "")}
            />
          </FormGroup>
          <FormGroup isStatic label="Duration">
            {this.state.duration}
          </FormGroup>
        </div>
      </Card>
    )
  }
}

export default RunStatusBar
