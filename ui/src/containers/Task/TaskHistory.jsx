import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import moment from 'moment'
import { has } from 'lodash'
import { renderModal, unrenderModal, stopRun } from '../../actions/'
import {
  serverTableConnect,
  Loader,
  ServerTable,
  StopRunModal,
  RunStatusText,
} from '../../components/'
import { getApiRoot, runStatusTypes } from '../../constants/'
import { calculateTaskDuration, getRunStatus } from '../../utils/'

const TaskHistoryRow = ({ path, run, onStopClick, index }) => {
  const enhancedStatus = getRunStatus({ status: run.status, exitCode: run.exit_code })
  return (
    <Link
      to={path}
      className="tr"
    >
      <div className="td overflow-ellipsis" style={{ flex: 1 }}>
        {has(run, 'started_at') ? moment(run.started_at).fromNow() : '-'}
      </div>
      <div
        className="td overflow-ellipsis"
        style={{ flex: 1 }}
      >
        <RunStatusText
          enhancedStatus={enhancedStatus}
          status={run.status}
          exitCode={run.exit_code}
        />
      </div>
      <div className="td overflow-ellipsis" style={{ flex: 1 }}>
        {calculateTaskDuration(run)}
      </div>
      <div className="td overflow-ellipsis" style={{ flex: 1 }}>
        {run.run_id}
      </div>
      <div className="td overflow-ellipsis" style={{ flex: 1 }}>
        {run.cluster}
      </div>
      <div className="td" style={{ flex: 1 }}>
        {
          (run.status === runStatusTypes.queued ||
          run.status === runStatusTypes.pending ||
          run.status === runStatusTypes.running) &&
            <button
              onClick={(evt) => {
                evt.preventDefault()
                evt.stopPropagation()
                onStopClick()
              }}
              className="button button-error"
            >Stop</button>
        }
      </div>
    </Link>
  )
}

class TaskHistory extends Component {
  componentWillReceiveProps(nextProps) {
    if (this.props.params.taskID !== nextProps.params.taskID) {
      this.props.forceRefetch()
    }
  }
  renderKillModal({ taskID, runID }) {
    const { dispatch, forceRefetch } = this.props
    const modal = (
      <StopRunModal
        stopRun={() => {
          dispatch(stopRun({ taskID, runID }, (res) => {
            if (!!res.terminated) {
              forceRefetch()
              dispatch(unrenderModal())
            }
          }))
        }}
        closeModal={() => { this.props.dispatch(unrenderModal()) }}
      />
    )
    this.props.dispatch(renderModal({ modal }))
  }
  render() {
    const {
      data,
      isFetching,
      forceRefetch,
    } = this.props

    const headers = [
      { displayName: 'Started At', key: 'started_at', sortable: true, style: { flex: 1 } },
      { displayName: 'Status', key: 'status', sortable: true, style: { flex: 1 } },
      { displayName: 'Duration', key: 'duration', sortable: false, style: { flex: 1 } },
      { displayName: 'Run ID', key: 'run_id', sortable: true, style: { flex: 1 } },
      { displayName: 'Cluster', key: 'cluster_name', sortable: true, style: { flex: 1 } },
      { displayName: 'Actions', key: 'actions', sortable: false, style: { flex: 1 } },
    ]

    return (
      <div className="section-container run-info">
        <div className="section-header">
          <div className="section-header-text">Task History</div>
        </div>
        <ServerTable
          headers={headers}
          viewSectionProps={{
            hasHeader: true,
            header: 'History',
            headerRight: (<button onClick={() => { forceRefetch() }}>Refresh</button>)
          }}
          tableName="Task History"
          {...this.props}
        >
          {
            isFetching ?
              <Loader containerStyle={{ height: 960 }} /> :
            data && data.history ? data.history.map((d, i) => (
              <TaskHistoryRow
                path={`/runs/${d.run_id}`}
                index={i}
                run={d}
                key={`task-history-row-${i}`}
                onStopClick={() => {
                  this.renderKillModal({
                    taskID: d.definition_id,
                    runID: d.run_id
                  })
                }}
              />
            )) : null
          }
        </ServerTable>
      </div>
    )
  }
}

export default connect()(
  serverTableConnect({
    initialQuery: {
      sort_by: 'started_at',
      order: 'desc',
    },
    urlRoot: props => `${getApiRoot()}/task/${props.params.taskID}/history?`
  })(TaskHistory)
)
