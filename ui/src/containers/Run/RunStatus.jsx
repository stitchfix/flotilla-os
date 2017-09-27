import React, { Component } from 'react'
import { connect } from 'react-redux'
import { has, isEmpty } from 'lodash'
import { FormGroup, RunStatusText } from '../../components/'
import { getRunStatus, calculateTaskDuration } from '../../utils/'

class RunStatus extends Component {
  state = {
    duration: '-'
  }
  componentDidMount() {
    this.renderedDuration = window.setInterval(this.calculateDuration.bind(this), 1000)
  }
  componentWillUnmount() {
    window.clearInterval(this.renderedDuration)
  }
  calculateDuration() {
    const { info } = this.props

    if (!isEmpty(info)) {
      if (has(info, 'started_at')) {
        const start = info.started_at
        let end

        if (has(info, 'finished_at')) {
          end = info.finished_at
          window.clearInterval(this.renderedDuration)
        } else {
          end = new Date()
        }

        this.setState({
          duration: calculateTaskDuration({
            started_at: start,
            finished_at: end,
          })
        })
      }
    }
  }
  render() {
    const { info } = this.props
    const enhancedStatus = has(info, 'status') ? getRunStatus({
      status: info.status,
      exitCode: has(info, 'exit_code') ? info.exit_code : null
    }) : '-'

    return (
      <div className="section-container">
        <div className="run-status">
          <FormGroup
            label="Run Status"
            isStatic
            style={{ borderBottom: 0 }}
          >
            <RunStatusText
              enhancedStatus={enhancedStatus}
              status={info.status}
              exitCode={info.exit_code}
            />
          </FormGroup>
          <FormGroup
            label="Duration"
            isStatic
            style={{ borderBottom: 0 }}
          >
            {this.state.duration}
          </FormGroup>
        </div>
      </div>
    )
  }
}

function mapStateToProps(state) {
  return ({ info: state.run.info })
}

export default connect(mapStateToProps)(RunStatus)
