import React, { Component } from 'react'
import PropTypes from 'prop-types'
import { connect } from 'react-redux'
import {
  stopRun,
  renderModal,
  unrenderModal,
} from '../../actions/'
import {
  ErrorView,
  StopRunModal
} from '../../components/'
import {
  RunNav,
  RunInfo,
  RunLogs,
  RunStatus
} from '../'

class Run extends Component {
  static propTypes = {
    dispatch: PropTypes.func,
    params: PropTypes.shape({
      runID: PropTypes.string
    }),
    run: PropTypes.object,
    router: PropTypes.shape({
      push: PropTypes.func,
    })
  }
  renderKillModal() {
    const { dispatch, run, params } = this.props

    const modal = (
      <StopRunModal
        stopRun={() => {
          dispatch(stopRun({
            taskID: run.info.definition_id,
            runID: params.runID
          }, () => {
            dispatch(unrenderModal())
            this.reset()
            this.props.router.push(`/tasks/${run.info.definition_id}`)
          }))
        }}
        closeModal={() => { dispatch(unrenderModal()) }}
      />
    )
    dispatch(renderModal({ modal }))
  }
  render() {
    const { run } = this.props

    let viewContent

    if (!!run.infoHasError && !!run.infoError) {
      viewContent = <ErrorView error={run.infoError} />
    } else {
      viewContent = (
        <div className="layout-detail sidebar-right">
          <div className="layout-detail-sidebar" style={{ order: 1 }}>
            <RunInfo />
          </div>
          <div className="layout-detail-content" style={{ order: 0 }}>
            <RunStatus />
            <RunLogs />
          </div>
        </div>
      )
    }

    return (
      <div className="view-container">
        <RunNav
          hasError={run.infoHasError}
          onStop={() => { this.renderKillModal() }}
          isStopped={!!run.info && !!run.info.status ? run.info.status.toLowerCase() === 'stopped' : false}
        />
        <div className="view">
          {viewContent}
        </div>
      </div>
    )
  }
}

const mapStateToProps = ({ run }) => ({ run })

export default connect(mapStateToProps)(Run)
