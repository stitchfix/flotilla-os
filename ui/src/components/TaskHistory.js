import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import { connect } from "react-redux"
import qs from "qs"
import { has, get, pickBy, identity } from "lodash"
import moment from "moment"
import EmptyTable from "./EmptyTable"
import EnhancedRunStatus from "./EnhancedRunStatus"
import PaginationButtons from "./PaginationButtons"
import SortHeader from "./SortHeader"
import StopRunModal from "./StopRunModal"
import withServerList from "./withServerList"
import modalActions from "../actions/modalActions"
import runStatusTypes from "../constants/runStatusTypes"
import getRunDuration from "../utils/getRunDuration"
import config from "../config"

const getUrl = id => `${config.FLOTILLA_API}/task/${id}/history`
const defaultQuery = {
  page: 1,
  sort_by: "started_at",
  order: "desc",
}

class TaskHistory extends Component {
  static displayName = "TaskHistory"
  static propTypes = {
    definitionId: PropTypes.string.isRequired,
    fetch: PropTypes.func.isRequired,
  }

  componentWillReceiveProps(nextProps) {
    if (this.props.definitionId !== nextProps.definitionId) {
      this.props.fetch(
        `${getUrl(this.props.definitionId)}?${qs.stringify(defaultQuery)}`
      )
    }
  }

  render() {
    const { isLoading, error, data, query, updateQuery, dispatch } = this.props

    let content = <EmptyTable isLoading />

    if (isLoading) {
      content = <EmptyTable isLoading />
    } else if (error) {
      const errorDisplay = error.toString() || "An error occurred."
      content = <EmptyTable title={errorDisplay} error />
    } else if (has(data, "history")) {
      if (Array.isArray(data.history) && data.history.length > 0) {
        content = data.history.map(d => (
          <Link
            className="pl-tr unstyled-link hoverable"
            to={`/runs/${d.run_id}`}
            key={d.run_id}
          >
            <div className="pl-td">
              {has(d, "started_at") ? moment(d.started_at).fromNow() : "-"}
            </div>
            <div className="pl-td">
              <EnhancedRunStatus
                status={get(d, "status")}
                exitCode={get(d, "exit_code")}
              />
            </div>
            <div className="pl-td pl-hide-small">{getRunDuration(d)}</div>
            <div className="pl-td">{d.run_id}</div>
            <div className="pl-td pl-hide-small">{d.cluster}</div>
            <div className="pl-td pl-hide-small">
              {get(d, "status") === runStatusTypes.pending ||
              get(d, "status") === runStatusTypes.queued ||
              get(d, "status") === runStatusTypes.running ? (
                <button
                  className="pl-button pl-intent-error"
                  onClick={evt => {
                    evt.preventDefault()
                    evt.stopPropagation()

                    dispatch(
                      modalActions.renderModal(
                        <StopRunModal
                          definitionId={d.definition_id}
                          runId={d.run_id}
                        />
                      )
                    )
                  }}
                >
                  Stop
                </button>
              ) : null}
            </div>
          </Link>
        ))
      } else {
        content = (
          <EmptyTable
            title="This task hasn't been run yet. Run it?"
            actions={
              <Link
                className="pl-button pl-intent-primary"
                to={`/tasks/${this.props.definitionId}/run`}
              >
                Run Task
              </Link>
            }
          />
        )
      }
    }

    return (
      <div className="pl-table pl-bordered">
        <div className="pl-tr">
          <SortHeader
            currentSortKey={query.sort_by}
            currentOrder={query.order}
            display="Started At"
            sortKey="started_at"
            updateQuery={updateQuery}
          />
          <SortHeader
            currentSortKey={query.sort_by}
            currentOrder={query.order}
            display="Status"
            sortKey="status"
            updateQuery={updateQuery}
          />
          <div className="pl-th pl-hide-small">Duration</div>
          <SortHeader
            currentSortKey={query.sort_by}
            currentOrder={query.order}
            display="Run ID"
            sortKey="run_id"
            updateQuery={updateQuery}
          />
          <SortHeader
            currentSortKey={query.sort_by}
            currentOrder={query.order}
            display="Cluster"
            sortKey="cluster"
            updateQuery={updateQuery}
            className="pl-hide-small"
          />
          <div className="pl-th pl-hide-small">Actions</div>
        </div>
        {content}
        <PaginationButtons
          total={get(data, "total", 20)}
          buttonCount={5}
          offset={query.offset}
          limit={query.limit}
          updateQuery={updateQuery}
          activeButtonClassName="pl-intent-primary"
          wrapperEl={
            <div className="table-footer flex ff-rn j-c a-c with-horizontal-child-margin" />
          }
        />
      </div>
    )
  }
}

export default withServerList({
  limit: 20,
  defaultQuery,
  getUrl: (props, query) => {
    const q = qs.stringify(pickBy(query, identity))
    return `${getUrl(props.definitionId)}?${q}`
  },
})(connect()(TaskHistory)).withHOCStack
