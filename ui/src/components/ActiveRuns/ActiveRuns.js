import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import Helmet from "react-helmet"
import { get } from "lodash"
import moment from "moment"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import ModalContext from "../Modal/ModalContext"
import StopRunModal from "../Modal/StopRunModal"
import Navigation from "../Navigation/Navigation"
import RunStatus from "../Run/RunStatus"
import Button from "../styled/Button"
import View from "../styled/View"
import SecondaryText from "../styled/SecondaryText"
import historyTableFilters from "../../helpers/historyTableFilters"
import runStatusTypes from "../../helpers/runStatusTypes"
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
          <title>Runs</title>
        </Helmet>
        <Navigation />
        <AsyncDataTable
          shouldContinuouslyFetch
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
          getItemKey={(item, i) => get(item, "run_id", i)}
          getItems={data => data.history}
          getTotal={data => data.total}
          filters={historyTableFilters}
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
  renderModal: PropTypes.func.isRequired,
}

ActiveRuns.defaultProps = {
  renderModal: () => {},
}

export default props => (
  <ModalContext.Consumer>
    {ctx => (
      <ActiveRuns
        {...props}
        renderModal={ctx.renderModal}
        unrenderModal={ctx.unrenderModal}
      />
    )}
  </ModalContext.Consumer>
)
