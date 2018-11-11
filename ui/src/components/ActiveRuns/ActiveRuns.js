import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import { connect } from "react-redux"
import Helmet from "react-helmet"
import { get } from "lodash"
import moment from "moment"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import { asyncDataTableFilterTypes } from "../AsyncDataTable/AsyncDataTableFilter"
import ModalContext from "../Modal/ModalContext"
import StopRunModal from "../Modal/StopRunModal"
import RunStatus from "../Run/RunStatus"
import Button from "../styled/Button"
import View from "../styled/View"
import ViewHeader from "../styled/ViewHeader"
import SecondaryText from "../styled/SecondaryText"
import runStatusTypes from "../../constants/runStatusTypes"
import api from "../../api"

class ActiveRuns extends Component {
  handleStopButtonClick = runData => {
    this.props.renderModal(
      <StopRunModal
        runID={runData.run_id}
        definitionID={runData.definition_id}
      />
    )
  }

  render() {
    return (
      <View>
        <Helmet>
          <title>Tasks</title>
        </Helmet>
        <ViewHeader
          title="Active Runs"
          actions={
            <Link className="pl-button pl-intent-primary" to="/generic">
              Run Generic Task
            </Link>
          }
        />
        <AsyncDataTable
          requestFn={api.getActiveRuns}
          shouldRequest={(prevProps, currProps) => false}
          columns={{
            stop: {
              allowSort: false,
              displayName: "Stop",
              render: item => {
                return (
                  <Button onClick={this.handleStopButtonClick.bind(this, item)}>
                    Stop
                  </Button>
                )
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
              width: 0.4,
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
              width: 1,
            },
            alias: {
              allowSort: false,
              displayName: "Alias",
              render: item => (
                <Link to={`/tasks/${item.definition_id}`}>
                  {get(item, "alias", item.definition_id)}
                </Link>
              ),
              width: 2.5,
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
          filters={{
            cluster_name: {
              displayName: "Cluster Name",
              type: asyncDataTableFilterTypes.SELECT,
              options: this.props.clusterOptions,
              description: "Search runs running on a specific cluster.",
            },
          }}
          initialQuery={{
            page: 1,
            sort_by: "started_at",
            order: "desc",
            status: [
              runStatusTypes.running,
              runStatusTypes.pending,
              runStatusTypes.queued,
            ],
          }}
          emptyTableTitle="No tasks are currently running."
        />
      </View>
    )
  }
}

ActiveRuns.propTypes = {
  clusterOptions: PropTypes.arrayOf(
    PropTypes.shape({ label: PropTypes.string, value: PropTypes.string })
  ),
  renderModal: PropTypes.func.isRequired,
}

ActiveRuns.defaultProps = {
  clusterOptions: [],
  renderModal: () => {},
}

const mapStateToProps = state => ({
  clusterOptions: get(state, "selectOpts.cluster", []),
})

export default connect(mapStateToProps)(props => (
  <ModalContext.Consumer>
    {ctx => (
      <ActiveRuns
        {...props}
        renderModal={ctx.renderModal}
        unrenderModal={ctx.unrenderModal}
      />
    )}
  </ModalContext.Consumer>
))
