import * as React from "react"
import { Link } from "react-router-dom"
import Helmet from "react-helmet"
import { get } from "lodash"
import moment from "moment"
import ModalContext from "../Modal/ModalContext"
import StopRunModal from "../Modal/StopRunModal"
import Navigation from "../Navigation/Navigation"
import RunStatus from "../Run/RunStatus"
import Button from "../styled/Button"
import View from "../styled/View"
import StyledField from "../styled/Field"
import SecondaryText from "../styled/SecondaryText"
import ListRequest, {
  IChildProps as IListRequestChildProps,
} from "../ListRequest/ListRequest"
import api from "../../api"
import { IFlotillaRun, flotillaRunStatuses } from "../../types"
import DataTable from "../DataTable/DataTable"
import ReactSelectWrapper from "../ReactSelectWrapper/ReactSelectWrapper"
import { stringToSelectOpt } from "../../helpers/reactSelectHelpers"

interface IProps extends IListRequestChildProps {
  renderModal: (modal: React.ReactNode) => void
}

class Runs extends React.Component<IProps> {
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
      <View>
        <Helmet>
          <title>Flotilla | Runs</title>
        </Helmet>
        <Navigation />
        <div
          style={{
            display: "flex",
            flexFlow: "row nowrap",
            justifyContent: "flex-start",
            alignItems: "flex-start",
            width: "100%",
          }}
        >
          <div style={{}}>
            <StyledField label="Alias" description="Filter by alias.">
              <ReactSelectWrapper
                isCreatable
                isMulti
                name="alias"
                onChange={(value: string | string[]) => {
                  updateSearch("alias", value)
                }}
                value={get(queryParams, "alias", "")}
              />
            </StyledField>
            <StyledField label="Status" description="Filter by status.">
              <ReactSelectWrapper
                isMulti
                name="status"
                onChange={(value: string | string[]) => {
                  updateSearch("status", value)
                }}
                value={get(queryParams, "status", "")}
                options={Object.values(flotillaRunStatuses)
                  .filter(
                    v =>
                      v !== flotillaRunStatuses.FAILED &&
                      v !== flotillaRunStatuses.SUCCESS &&
                      v !== flotillaRunStatuses.STOPPED &&
                      v !== flotillaRunStatuses.NEEDS_RETRY
                  )
                  .map(stringToSelectOpt)}
              />
            </StyledField>
            <StyledField label="Cluster" description="Filter by cluster.">
              <ReactSelectWrapper
                name="cluster_name"
                onChange={(value: string | string[]) => {
                  updateSearch("cluster_name", value)
                }}
                value={get(queryParams, "cluster_name", "")}
                shouldRequestOptions
                requestOptionsFn={api.getClusters}
              />
            </StyledField>
          </div>
          <DataTable
            items={get(data, "history", [])}
            columns={{
              stop: {
                allowSort: false,
                displayName: "Stop",
                render: item => {
                  return (
                    <Button
                      onClick={this.handleStopButtonClick.bind(this, item)}
                    >
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
            onSortableHeaderClick={updateSort}
            getItemKey={(item, i) => get(item, "run_id", i)}
            currentSortKey={currentSortKey}
            currentSortOrder={currentSortOrder}
            currentPage={currentPage}
          />
        </div>
      </View>
    )
  }
}

const ConnectedRuns: React.FunctionComponent = () => (
  <ModalContext.Consumer>
    {ctx => (
      <ListRequest
        getRequestArgs={query => ({ query })}
        initialQuery={{
          page: 1,
          sort_by: "started_at",
          order: "desc",
          status: [
            flotillaRunStatuses.RUNNING,
            flotillaRunStatuses.PENDING,
            flotillaRunStatuses.QUEUED,
          ],
        }}
        limit={50}
        requestFn={api.getActiveRuns}
        shouldContinuouslyFetch={false}
      >
        {(props: IListRequestChildProps) => (
          <Runs {...props} renderModal={ctx.renderModal} />
        )}
      </ListRequest>
    )}
  </ModalContext.Consumer>
)

export default ConnectedRuns
