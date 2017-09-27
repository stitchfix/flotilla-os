import React, { Component } from 'react'
import { connect } from 'react-redux'
import { RefreshCw, X, Layout } from 'react-feather'
import ReactTooltip from 'react-tooltip'
import { has, takeRight } from 'lodash'
import Ansi from 'ansi-to-react'
import { stopRun } from '../../actions/'
import {
  RunStatus,
} from '../'

class RunMini extends Component {
  state = {
    renderStopConfirmation: false
  }
  constructor(props) {
    super(props)
    this.onStopClick = this.onStopClick.bind(this)
    this.onRetryClick = this.onRetryClick.bind(this)
    this.returnToMainClick = this.returnToMainClick.bind(this)
  }
  onStopClick() {
    const { dispatch, runInfo, params } = this.props
    const { renderStopConfirmation } = this.state
    if (!!renderStopConfirmation) {
      dispatch(stopRun({
        taskID: runInfo.definition_id,
        runID: params.runID
      }))
    } else {
      this.setState({ renderStopConfirmation: true })
    }
  }
  onRetryClick() {
    const { runInfo } = this.props
    const url = `${window.location.origin}/#/tasks/${runInfo.definition_id}/run`
    window.open(url, '_blank')
  }
  returnToMainClick() {
    const { runInfo } = this.props
    const url = `${window.location.origin}/#/runs/${runInfo.run_id}`
    window.open(url, '_blank')
  }
  render() {
    const { runInfo, mostRecentLogs } = this.props
    const { renderStopConfirmation } = this.state
    const isStopped = !!runInfo && !!runInfo.status ?
      runInfo.status.toLowerCase() === 'stopped' :
      false

    return (
      <div style={{ padding: 12 }}>
        <div
          className="flex ff-rn j-sb a-c"
          style={{ marginBottom: 12, paddingLeft: 12 }}
        >
          <div className="overflow-ellipsis">
            <a onClick={this.returnToMainClick}>{has(runInfo, 'run_id') && runInfo.run_id}</a>
          </div>
          <div className="flex">
            <ReactTooltip id="stopButton" effect="solid">
              Stop Run
            </ReactTooltip>
            {
              !isStopped &&
                <button
                  data-tip
                  data-for="stopButton"
                  className="button button-small button-error"
                  onClick={this.onStopClick}
                >
                  {renderStopConfirmation ? 'Confirm' : <X size={14} />}
                </button>
            }
            <ReactTooltip id="retryButton" effect="solid">
              Retry
            </ReactTooltip>
            <button
              data-tip
              data-for="retryButton"
              className="button button-small"
              onClick={this.onRetryClick}
            >
              <RefreshCw size={14} />
            </button>
            <ReactTooltip id="detailButton" effect="solid">
              Detailed View
            </ReactTooltip>
            <button
              data-tip
              data-for="detailButton"
              className="button button-small"
              onClick={this.returnToMainClick}
            >
              <Layout size={14} />
            </button>
          </div>
        </div>
        <RunStatus />
        {
          // <div className="section-container">
          //   <div className="section-content code">
          //     <pre>
          //       {
          //         !!mostRecentLogs && mostRecentLogs
          //           .map((line, i) => (
          //             <div key={`logs-line-${i}`}><Ansi>{line}</Ansi></div>
          //           ))
          //       }
          //     </pre>
          //   </div>
          // </div>
        }
      </div>
    )
  }
}

const mapStateToProps = ({ run }) => {
  const runInfo = run.info
  // const mostRecentLogs = has(run, 'logs') && run.logs.length > 0 ? takeRight(run.logs[run.logs.length - 1].logs, 5) : null
  return ({
    runInfo,
    // mostRecentLogs,
  })
}

export default connect(mapStateToProps)(RunMini)
