import React, { Component } from 'react'
import { connect } from 'react-redux'
import { Link } from 'react-router'
import moment from 'moment'
import { stopRun, renderModal, unrenderModal } from '../../actions/'
import {
  ReactSelectWrapper,
  serverTableConnect,
  Loader,
  ServerTable,
  StopRunModal,
  RunStatusText
} from '../../components/'
import { getApiRoot } from '../../constants/'
import { getRunStatus } from '../../utils/'

const RunRow = ({ run, pathname, renderStopModal, index }) => {
  const enhancedStatus = getRunStatus({ status: run.status, exitCode: run.exit_code })
  return (
    <Link
      to={{ pathname }}
      className="tr"
    >
      <div className="td flex j-c" style={{ flex: 1 }}>
        <button
          onClick={(evt) => {
            evt.preventDefault()
            evt.stopPropagation()
            renderStopModal()
          }}
          className="button"
        >Stop</button>
      </div>
      <div className="td" style={{ flex: 1 }}>
        <RunStatusText
          enhancedStatus={enhancedStatus}
          status={run.status}
          exitCode={run.exit_code}
        />
      </div>
      <div className="td" style={{ flex: 3 }}>
        {moment(run.started_at).fromNow()}
      </div>
      <div className="td" style={{ flex: 6 }}>
        {run.alias || run.definition_id}
      </div>
      <div className="td" style={{ flex: 2 }}>
        {run.cluster}
      </div>
    </Link>
  )
}

class Runs extends Component {
  renderStopModal({ taskID, runID }) {
    const { dispatch, forceRefetch } = this.props
    const modal = (
      <StopRunModal
        stopRun={() => {
          dispatch(stopRun({ taskID, runID }, (res) => {
            if (!!res.stopped) {
              forceRefetch()
              dispatch(unrenderModal())
            }
          }))
        }}
        closeModal={() => { dispatch(unrenderModal()) }}
      />
    )
    dispatch(renderModal({ modal }))
  }
  render() {
    const {
      data,
      onQueryChange,
      query,
      clusterOpts,
      isFetching,
    } = this.props
    const headers = [
      { displayName: 'Actions', key: 'actions', sortable: false, style: { flex: 1, justifyContent: 'center' } },
      { displayName: 'Status', key: 'status', sortable: true, style: { flex: 1, justifyContent: 'center' } },
      { displayName: 'Started At', key: 'started_at', sortable: true, style: { flex: 3 } },
      { displayName: 'Alias', key: 'alias', sortable: false, style: { flex: 6 } },
      { displayName: 'Cluster', key: 'cluster_name', sortable: true, style: { flex: 2 } },
    ]
    const queryInputs = [
      {
        style: {},
        label: 'Cluster',
        input: (
          <ReactSelectWrapper
            value={query.cluster_name}
            onChange={(o) => { onQueryChange('cluster_name', o) }}
            options={clusterOpts}
          />
        )
      }
    ]
    return (
      <div className="section-container">
        <ServerTable
          headers={headers}
          queryInputs={queryInputs}
          {...this.props}
        >
          {
            isFetching ?
              <Loader containerStyle={{ height: 960 }} /> :
            data && data.history ? data.history.map((run, i) => (
              <RunRow
                key={`run-row-${i}`}
                run={run}
                index={i}
                pathname={`/runs/${run.run_id}`}
                renderStopModal={() => {
                  this.renderStopModal({
                    taskID: run.definition_id,
                    runID: run.run_id
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

const mapStateToProps = state => ({
  clusterOpts: state.dropdownOpts.cluster,
})

export default connect(mapStateToProps)(
  serverTableConnect({
    urlRoot: () => `${getApiRoot()}/history?status=RUNNING&status=PENDING&status=QUEUED&`,
    initialQuery: {
      sort_by: 'started_at',
      order: 'desc',
    }
  })(Runs)
)
