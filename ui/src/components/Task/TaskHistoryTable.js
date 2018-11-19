import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import moment from "moment"
import { get, omit } from "lodash"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import { asyncDataTableFilterTypes } from "../AsyncDataTable/AsyncDataTableFilter"
import api from "../../api"
import RunStatus from "../Run/RunStatus"
import Button from "../styled/Button"
import ButtonLink from "../styled/ButtonLink"
import SecondaryText from "../styled/SecondaryText"
import runStatusTypes from "../../constants/runStatusTypes"
import getRunDuration from "../../utils/getRunDuration"
import StopRunModal from "../Modal/StopRunModal"
import ModalContext from "../Modal/ModalContext"
import historyTableFilters from "../../utils/historyTableFilters"

class TaskHistoryTable extends Component {
  static isTaskActive = status =>
    status === runStatusTypes.pending ||
    status === runStatusTypes.queued ||
    status === runStatusTypes.running

  handleStopButtonClick = runData => {
    this.props.renderModal(
      <StopRunModal
        runID={runData.run_id}
        definitionID={runData.definition_id}
      />
    )
  }

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
            displayName: "Stop",
            render: item => {
              if (TaskHistoryTable.isTaskActive(item.status)) {
                return (
                  <Button onClick={this.handleStopButtonClick.bind(this, item)}>
                    Stop
                  </Button>
                )
              }

              return null
            },
            width: 0.6,
          },
          status: {
            allowSort: true,
            displayName: "Status",
            render: item => (
              <RunStatus
                status={get(item, "status")}
                exitCode={get(item, "exit_code")}
              />
            ),
            width: 0.2,
          },
          started_at: {
            allowSort: true,
            displayName: "Started At",
            render: item => {
              if (!!get(item, "started_at")) {
                return (
                  <div>
                    <div style={{ marginBottom: 4 }}>
                      {moment(item.started_at).fromNow()}
                    </div>
                    <SecondaryText>{item.started_at}</SecondaryText>
                  </div>
                )
              }
              return "-"
            },
            width: 0.8,
          },
          duration: {
            allowSort: false,
            displayName: "Duration",
            render: item => getRunDuration(item),
            width: 0.5,
          },
          run_id: {
            allowSort: true,
            displayName: "Run ID",
            render: item => (
              <Link to={`/runs/${item.run_id}`}>{item.run_id}</Link>
            ),
            width: 1,
          },
          cluster: {
            allowSort: false,
            displayName: "Cluster",
            render: item => item.cluster,
            width: 1,
          },
        }}
        getItems={data => data.history}
        getTotal={data => data.total}
        filters={omit(historyTableFilters, ["alias"])}
        initialQuery={{
          page: 1,
          sort_by: "started_at",
          order: "desc",
        }}
        emptyTableTitle="This task hasn't been run yet."
        emptyTableBody={
          <ButtonLink to={`/tasks/${definitionID}/run`}>Run Task</ButtonLink>
        }
        isView={false}
      />
    )
  }
}

TaskHistoryTable.propTypes = {
  definitionID: PropTypes.string.isRequired,
  renderModal: PropTypes.func.isRequired,
}

TaskHistoryTable.defaultProps = {
  renderModal: () => {},
}

export default props => (
  <ModalContext.Consumer>
    {ctx => <TaskHistoryTable {...props} renderModal={ctx.renderModal} />}
  </ModalContext.Consumer>
)
