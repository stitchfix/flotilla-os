import qs from "qs"
import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import moment from "moment"
import { get } from "lodash"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import { asyncDataTableFilterTypes } from "../AsyncDataTable/AsyncDataTableFilter"
import api from "../../api"
import config from "../../config"
import EnhancedRunStatus from "../EnhancedRunStatus"
import Button from "../Button"
import runStatusTypes from "../../constants/runStatusTypes"
import getRunDuration from "../../utils/getRunDuration"

class TaskHistoryTable extends Component {
  static isTaskActive = status =>
    status === runStatusTypes.pending ||
    status === runStatusTypes.queued ||
    status === runStatusTypes.running

  render() {
    const { definitionID } = this.props

    return (
      <AsyncDataTable
        getRequestArgs={query => ({
          definitionID,
          query,
        })}
        requestFn={api.getTaskHistory}
        shouldRequest={(prevProps, currProps) =>
          prevProps.definitionID !== currProps.definitionID
        }
        columns={{
          stop: {
            allowSort: false,
            displayName: "Stop Run",
            render: item => {
              if (TaskHistoryTable.isTaskActive(item.status)) {
                return <Button>Stop</Button>
              }

              return null
            },
          },
          status: {
            allowSort: true,
            displayName: "Status",
            render: item => (
              <EnhancedRunStatus
                status={get(item, "status")}
                exitCode={get(item, "exit_code")}
              />
            ),
          },
          started_at: {
            allowSort: true,
            displayName: "Started At",
            render: item => moment(item.started_at).fromNow(),
            width: 1,
          },
          duration: {
            allowSort: false,
            displayName: "Duration",
            render: item => getRunDuration(item),
          },
          run_id: {
            allowSort: true,
            displayName: "Run ID",
            render: item => (
              <Link to={`/runs/${item.run_id}`}>{item.run_id}</Link>
            ),
          },
          cluster: {
            allowSort: false,
            displayName: "Cluster",
            render: item => item.cluster,
          },
        }}
        getItems={data => data.history}
        getTotal={data => data.total}
        filters={{}}
        initialQuery={{
          page: 1,
          sort_by: "started_at",
          order: "desc",
        }}
        emptyTableTitle="This task hasn't been run yet."
        emptyTableBody={
          <Link className="pl-button pl-intent-primary" to="/tasks/create">
            TODO: make this
          </Link>
        }
      />
    )
  }
}

TaskHistoryTable.propTypes = {
  definitionID: PropTypes.string.isRequired,
}

TaskHistoryTable.defaultProps = {}

export default TaskHistoryTable
