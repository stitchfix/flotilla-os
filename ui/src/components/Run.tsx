import * as React from "react"
import { get } from "lodash"
import { Link, RouteComponentProps } from "react-router-dom"
import {
  Card,
  Spinner,
  Classes,
  ButtonGroup,
  Button,
  Collapse,
  Pre,
  Icon,
  Tag,
  Tabs,
  Tab,
  Tooltip,
} from "@blueprintjs/core"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import api from "../api"
import { Run as RunShape, RunStatus, ExecutionEngine, RunTabId } from "../types"
import Attribute from "./Attribute"
import EnvList from "./EnvList"
import ViewHeader from "./ViewHeader"
import StopRunButton from "./StopRunButton"
import { RUN_FETCH_INTERVAL_MS } from "../constants"
import Toggler from "./Toggler"
import ISO8601AttributeValue from "./ISO8601AttributeValue"
import LogRequester from "./LogRequester"
import RunTag from "./RunTag"
import Duration from "./Duration"
import RunEvents from "./RunEvents"
import QueryParams, { ChildProps as QPChildProps } from "./QueryParams"
import { RUN_TAB_ID_QUERY_KEY } from "../constants"

export type Props = QPChildProps &
  RequestChildProps<RunShape, { runID: string }> & {
    runID: string
  }

type State = {
  hasLogs: boolean
}

export class Run extends React.Component<Props, State> {
  requestIntervalID: number | undefined

  constructor(props: Props) {
    super(props)
    this.request = this.request.bind(this)
    this.setHasLogs = this.setHasLogs.bind(this)
  }

  state = {
    hasLogs: false,
  }

  componentDidMount() {
    const { data } = this.props

    // If data has been fetched and the run hasn't stopped, start polling.
    if (data && data.status !== RunStatus.STOPPED) this.setRequestInterval()
  }

  componentDidUpdate(prevProps: Props) {
    if (
      prevProps.requestStatus === RequestStatus.NOT_READY &&
      this.props.requestStatus === RequestStatus.READY &&
      this.props.data &&
      this.props.data.status !== RunStatus.STOPPED
    ) {
      // If the RequestStatus transitions from NOT_READY to READY and the run
      // isn't stopped, start polling.
      this.setRequestInterval()
    }

    if (this.props.data && this.props.data.status === RunStatus.STOPPED) {
      // If the Run transitions to a STOPPED state, stop polling.
      this.clearRequestInterval()
    }
  }

  componentWillUnmount() {
    if (this.request !== undefined) this.clearRequestInterval()
  }

  request() {
    const { isLoading, error, request, runID } = this.props
    if (isLoading === true || error !== null) return
    request({ runID })
  }

  setRequestInterval() {
    this.requestIntervalID = window.setInterval(
      this.request,
      RUN_FETCH_INTERVAL_MS
    )
  }

  clearRequestInterval() {
    window.clearInterval(this.requestIntervalID)
  }

  getLogsHeight(): number {
    if (window.innerWidth >= 1230) {
      return window.innerHeight - 78 - 50 - 24 - 30 - 20
    }

    return 720
  }

  getActiveTabId(): RunTabId {
    const { data, query } = this.props
    const { hasLogs } = this.state
    const queryTabId: RunTabId | null = get(query, RUN_TAB_ID_QUERY_KEY, null)

    if (queryTabId === null) {
      if (hasLogs === true) {
        return RunTabId.LOGS
      }

      if (
        data &&
        data.engine === ExecutionEngine.EKS &&
        data.status !== RunStatus.STOPPED
      ) {
        return RunTabId.EVENTS
      }

      return RunTabId.LOGS
    }

    return queryTabId
  }

  setActiveTabId(id: RunTabId): void {
    this.props.setQuery({ [RUN_TAB_ID_QUERY_KEY]: id })
  }

  setHasLogs() {
    if (this.state.hasLogs === false) {
      this.setState({ hasLogs: true })
    }
  }

  render() {
    const { data, requestStatus, runID } = this.props

    if (requestStatus === RequestStatus.READY && data) {
      let btn: React.ReactNode = null

      if (data.status === RunStatus.STOPPED) {
        btn = (
          <Link
            className={Classes.BUTTON}
            to={{
              pathname: `/tasks/${data.definition_id}/execute`,
              state: data,
            }}
          >
            <div className="bp3-button-text">Retry</div>
            <Icon icon="repeat" />
          </Link>
        )
      } else {
        btn = <StopRunButton runID={runID} definitionID={data.definition_id} />
      }

      return (
        <Toggler>
          {metadataVisibility => (
            <>
              <ViewHeader
                leftButton={
                  <Button
                    onClick={metadataVisibility.toggleVisibility}
                    icon={
                      metadataVisibility.isVisible ? "menu-closed" : "menu-open"
                    }
                    style={{ marginRight: 12 }}
                  >
                    {metadataVisibility.isVisible ? "Hide" : "Show"}
                  </Button>
                }
                breadcrumbs={[
                  {
                    text: data.alias,
                    href: `/tasks/${data.definition_id}`,
                  },
                  {
                    text: data.run_id,
                    href: `/runs/${data.run_id}`,
                  },
                ]}
                buttons={btn}
              />
              <div className="flotilla-sidebar-view-container">
                {metadataVisibility.isVisible && (
                  <div className="flotilla-sidebar-view-sidebar">
                    <Card style={{ marginBottom: 12 }}>
                      <div
                        className="flotilla-attributes-container flotilla-attributes-container-horizontal"
                        style={{ marginBottom: 12 }}
                      >
                        <Attribute name="Status" value={<RunTag {...data} />} />
                        <Attribute
                          name="Duration"
                          value={
                            data.started_at && (
                              <Duration
                                start={data.started_at}
                                end={data.finished_at}
                                isActive={data.status !== RunStatus.STOPPED}
                              />
                            )
                          }
                        />
                        <Attribute name="" value="" />
                      </div>
                      <div className="flotilla-form-section-divider" />
                      <div
                        className="flotilla-attributes-container flotilla-attributes-container-horizontal"
                        style={{ marginBottom: 12 }}
                      >
                        <Attribute name="Exit Code" value={data.exit_code} />
                        <Attribute
                          name="Exit Reason"
                          value={data.exit_reason}
                          containerStyle={{ flex: "2 2" }}
                        />
                      </div>
                      <div className="flotilla-form-section-divider" />
                      <div
                        className="flotilla-attributes-container flotilla-attributes-container-horizontal"
                        style={{ marginBottom: 12 }}
                      >
                        <Attribute
                          name="Engine Type"
                          value={<Tag>{data.engine}</Tag>}
                        />
                        <Attribute name="Cluster" value={data.cluster} />

                        <Attribute
                          name="Node Lifecycle"
                          value={<Tag>{data.node_lifecycle || "-"}</Tag>}
                        />
                      </div>
                      <div className="flotilla-form-section-divider" />
                      <div
                        className="flotilla-attributes-container flotilla-attributes-container-horizontal"
                        style={{ marginBottom: 12 }}
                      >
                        <Attribute name="CPU (Units)" value={data.cpu} />
                        <Attribute name="Memory (MB)" value={data.memory} />
                        <Attribute
                          name="Disk Size (GB)"
                          value={data.ephemeral_storage || "-"}
                        />
                        <Attribute name="GPU Count" value={data.gpu || 0} />
                      </div>
                      <div className="flotilla-form-section-divider" />
                      <div
                        className="flotilla-attributes-container flotilla-attributes-container-horizontal"
                        style={{ marginBottom: 12 }}
                      >
                        <Attribute
                          name="Queued At"
                          value={
                            <ISO8601AttributeValue time={data.queued_at} />
                          }
                        />
                        <Attribute
                          name="Started At"
                          value={
                            <ISO8601AttributeValue time={data.started_at} />
                          }
                        />
                        <Attribute
                          name="Finished At"
                          value={
                            <ISO8601AttributeValue time={data.finished_at} />
                          }
                        />
                      </div>
                      <div className="flotilla-form-section-divider" />
                      <div className="flotilla-attributes-container flotilla-attributes-container-vertical">
                        <Attribute name="Run ID" value={data.run_id} />
                        <Attribute
                          name="Definition ID"
                          value={data.definition_id}
                        />
                        <Attribute name="Image" value={data.image} />
                        <Attribute
                          name="Command"
                          value={
                            data.command ? (
                              <Pre className="flotilla-pre">
                                {data.command.replace(/\n(\s)+/g, "\n")}
                              </Pre>
                            ) : (
                              "-"
                            )
                          }
                        />
                      </div>
                    </Card>
                    <Card>
                      <div className="flotilla-card-header-container">
                        <div className="flotilla-card-header">
                          Environment Variables
                        </div>
                      </div>
                      <EnvList env={data.env} />
                    </Card>
                  </div>
                )}
                <div className="flotilla-sidebar-view-content">
                  <Tabs
                    selectedTabId={this.getActiveTabId()}
                    onChange={id => {
                      this.setActiveTabId(id as RunTabId)
                    }}
                  >
                    <Tab
                      id={RunTabId.LOGS}
                      title="Container Logs"
                      panel={
                        <LogRequester
                          runID={data.run_id}
                          status={data.status}
                          height={this.getLogsHeight()}
                          setHasLogs={this.setHasLogs}
                        />
                      }
                    />
                    <Tab
                      id={RunTabId.EVENTS}
                      title={
                        data.engine !== ExecutionEngine.EKS ? (
                          <Tooltip content="Run events are only available for tasks run on EKS.">
                            EKS Pod Events
                          </Tooltip>
                        ) : (
                          "EKS Pod Events"
                        )
                      }
                      panel={
                        <RunEvents
                          runID={data.run_id}
                          status={data.status}
                          hasLogs={this.state.hasLogs}
                        />
                      }
                      disabled={data.engine !== ExecutionEngine.EKS}
                    />
                  </Tabs>
                </div>
              </div>
            </>
          )}
        </Toggler>
      )
    }

    if (requestStatus === RequestStatus.ERROR) return <div>errro</div>
    return <Spinner />
  }
}

const Connected: React.FunctionComponent<RouteComponentProps<{
  runID: string
}>> = ({ match }) => (
  <QueryParams>
    {({ query, setQuery }) => (
      <Request<RunShape, { runID: string }>
        requestFn={api.getRun}
        initialRequestArgs={{ runID: match.params.runID }}
      >
        {props => (
          <Run
            {...props}
            runID={match.params.runID}
            query={query}
            setQuery={setQuery}
          />
        )}
      </Request>
    )}
  </QueryParams>
)

export default Connected
