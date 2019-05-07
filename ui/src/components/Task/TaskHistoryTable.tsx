import * as React from "react"
import { Link } from "react-router-dom"
import moment from "moment"
import { get } from "lodash"
import api from "../../api"
import RunStatus from "../Run/RunStatus"
import Button from "../styled/Button"
import SecondaryText from "../styled/SecondaryText"
import getRunDuration from "../../helpers/getRunDuration"
import StopRunModal from "../Modal/StopRunModal"
import ModalContext from "../Modal/ModalContext"
import DataTable from "../DataTable/DataTable"
import ListRequest, {
  IChildProps as IListRequestChildProps,
} from "../ListRequest/ListRequest"
import { flotillaRunStatuses, IFlotillaRun } from "../../types"

interface IProps extends IListRequestChildProps {
  definitionID: string
  renderModal: (modal: React.ReactNode) => void
}

class TaskHistoryTable extends React.PureComponent<IProps> {
  static isRunActive = (status: flotillaRunStatuses): boolean =>
    status === flotillaRunStatuses.PENDING ||
    status === flotillaRunStatuses.QUEUED ||
    status === flotillaRunStatuses.RUNNING

  handleStopButtonClick = (runData: IFlotillaRun): void => {
    this.props.renderModal(
      <StopRunModal
        runID={runData.run_id}
        definitionID={runData.definition_id}
      />
    )
  }

  render() {
    const {
      queryParams,
      updateSearch,
      data,
      updateSort,
      currentSortKey,
      currentSortOrder,
      currentPage,
    } = this.props

    return (
      <DataTable
        items={get(data, "history", [])}
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
          duration: {
            allowSort: false,
            displayName: "Duration",
            render: item => getRunDuration(item),
            width: 1,
          },
        }}
        onSortableHeaderClick={updateSort}
        getItemKey={(item, i) => get(item, "run_id", i)}
        currentSortKey={currentSortKey}
        currentSortOrder={currentSortOrder}
        currentPage={currentPage}
      />
    )
  }
}

interface IConnectedProps {
  definitionID: string
}

const ConnectedTaskHistoryTable: React.FunctionComponent<IConnectedProps> = ({
  definitionID,
}) => (
  <ModalContext.Consumer>
    {ctx => (
      <ListRequest
        getRequestArgs={query => ({ query, definitionID: definitionID })}
        initialQuery={{
          page: 1,
          sort_by: "started_at",
          order: "desc",
        }}
        limit={50}
        requestFn={api.getTaskHistory}
        shouldContinuouslyFetch={false}
      >
        {(props: IListRequestChildProps) => (
          <TaskHistoryTable
            {...props}
            definitionID={definitionID}
            renderModal={ctx.renderModal}
          />
        )}
      </ListRequest>
    )}
  </ModalContext.Consumer>
)

export default ConnectedTaskHistoryTable
