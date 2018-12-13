import * as React from "react"
import { Link } from "react-router-dom"
import moment from "moment"
import { get, omit } from "lodash"
import AsyncDataTable from "../AsyncDataTable/AsyncDataTable"
import api from "../../api"
import RunStatus from "../Run/RunStatus"
import Button from "../styled/Button"
import ButtonLink from "../styled/ButtonLink"
import SecondaryText from "../styled/SecondaryText"
import getRunDuration from "../../helpers/getRunDuration"
import StopRunModal from "../Modal/StopRunModal"
import ModalContext from "../Modal/ModalContext"
import historyTableFilters from "../../helpers/historyTableFilters"
import { ecsRunStatuses, IFlotillaRun, intents } from "../../.."

interface IUnwrappedTaskHistoryTableProps {
  definitionID: string
}

interface ITaskHistoryTableProps extends IUnwrappedTaskHistoryTableProps {
  renderModal: (modal: React.ReactNode) => void
}

class TaskHistoryTable extends React.PureComponent<ITaskHistoryTableProps> {
  static isTaskActive = (status: ecsRunStatuses): boolean =>
    status === ecsRunStatuses.PENDING ||
    status === ecsRunStatuses.QUEUED ||
    status === ecsRunStatuses.RUNNING

  handleStopButtonClick = (runData: IFlotillaRun): void => {
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
        limit={50}
        shouldContinuouslyFetch={false}
        getItemKey={(item: any, index: number) => index}
        getRequestArgs={query => ({
          definitionID,
          query,
        })}
        requestFn={api.getTaskHistory}
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
        emptyTableTitle="No items were found."
        emptyTableBody={
          <ButtonLink
            intent={intents.PRIMARY}
            to={`/tasks/${definitionID}/run`}
          >
            Run Task
          </ButtonLink>
        }
        isView={false}
      />
    )
  }
}

const WrappedTaskHistoryTable: React.SFC<
  IUnwrappedTaskHistoryTableProps
> = props => (
  <ModalContext.Consumer>
    {ctx => <TaskHistoryTable {...props} renderModal={ctx.renderModal} />}
  </ModalContext.Consumer>
)

export default WrappedTaskHistoryTable
