import * as React from "react"
import { connect, ConnectedProps } from "react-redux"
import { get } from "lodash"
import { Link, RouteComponentProps } from "react-router-dom"
import {
  Card,
  Spinner,
  Classes,
  Button,
  Icon,
  Tabs,
  Tab,
  Tooltip,
  Callout,
  Intent,
} from "@blueprintjs/core"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import api from "../api"
import {
  Run as RunShape,
  RunStatus,
  ExecutionEngine,
  RunTabId,
  ExecutableType,
  EnhancedRunStatusEmojiMap,
  EnhancedRunStatus,
} from "../types"
import EnvList from "./EnvList"
import ViewHeader from "./ViewHeader"
import StopRunButton from "./StopRunButton"
import { RUN_FETCH_INTERVAL_MS } from "../constants"
import Toggler from "./Toggler"
import LogRequesterCloudWatchLogs from "./LogRequesterCloudWatchLogs"
import LogRequesterS3 from "./LogRequesterS3"
import RunEvents from "./RunEvents"
import QueryParams, { ChildProps as QPChildProps } from "./QueryParams"
import { RUN_TAB_ID_QUERY_KEY } from "../constants"
import Attribute from "./Attribute"
import RunTag from "./RunTag"
import Duration from "./Duration"
import ErrorCallout from "./ErrorCallout"
import RunSidebar from "./RunSidebar"
import Helmet from "react-helmet"
import AutoscrollSwitch from "./AutoscrollSwitch"
import { RootState } from "../state/store"
import RecursiveKeyValueRenderer from "./RecursiveKeyValueRenderer"
import CloudtrailRecords from "./CloudtrailRecords"
import getEnhancedRunStatus from "../helpers/getEnhancedRunStatus"

const connected = connect((state: RootState) => state.runView)

export type Props = QPChildProps &
  RequestChildProps<RunShape, { runID: string }> & {
    runID: string
  } & ConnectedProps<typeof connected>

export class Run extends React.Component<Props> {
  requestIntervalID: number | undefined

  constructor(props: Props) {
    super(props)
    this.request = this.request.bind(this)
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
    window.clearInterval(this.requestIntervalID)
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

  getActiveTabId(): RunTabId {
    const { data, query, hasLogs } = this.props
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

  getExecutableLinkName(): string {
    const { data } = this.props
    if (data) {
      switch (data.executable_type) {
        case ExecutableType.ExecutableTypeDefinition:
          return data.alias
        case ExecutableType.ExecutableTypeTemplate:
          return data.executable_id
      }
    }
    return ""
  }

  getExecutableLinkURL(): string {
    const { data } = this.props
    if (data) {
      switch (data.executable_type) {
        case ExecutableType.ExecutableTypeDefinition:
          return `/tasks/${data.definition_id}`
        case ExecutableType.ExecutableTypeTemplate:
          return `/templates/${data.executable_id}`
      }
    }
    return ""
  }

  render() {
    const { data, requestStatus, runID, error } = this.props

    switch (requestStatus) {
      case RequestStatus.ERROR:
        return <ErrorCallout error={error} />
      case RequestStatus.READY:
        if (data) {
          const cloudtrailRecords = get(
            data,
            ["cloudtrail_notifications", "Records"],
            null
          )
          const hasCloudtrailRecords = cloudtrailRecords !== null
          let btn: React.ReactNode = null

          if (data.status === RunStatus.STOPPED) {
            btn = (
              <Link
                className={Classes.BUTTON}
                to={{
                  pathname: `${this.getExecutableLinkURL()}/execute`,
                  state: data,
                }}
              >
                <div className="bp3-button-text">Retry</div>
                <Icon icon="repeat" />
              </Link>
            )
          } else {
            btn = (
              <StopRunButton runID={runID} definitionID={data.definition_id} />
            )
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
                          metadataVisibility.isVisible
                            ? "menu-closed"
                            : "menu-open"
                        }
                        style={{ marginRight: 12 }}
                      >
                        {metadataVisibility.isVisible ? "Hide" : "Show"}
                      </Button>
                    }
                    breadcrumbs={[
                      {
                        text: this.getExecutableLinkName(),
                        href: this.getExecutableLinkURL(),
                      },
                      {
                        text: data.run_id,
                        href: `/runs/${data.run_id}`,
                      },
                    ]}
                    buttons={btn}
                  />
                  <div className="flotilla-sidebar-view-container">
                    {metadataVisibility.isVisible && <RunSidebar data={data} />}
                    <div className="flotilla-sidebar-view-content">
                      <Card style={{ marginBottom: 12 }}>
                        <div className="flotilla-attributes-container flotilla-attributes-container-horizontal">
                          <Attribute
                            name="Status"
                            value={<RunTag {...data} />}
                          />
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
                          <Attribute name="Exit Code" value={data.exit_code} />
                          <Attribute
                            name="Exit Reason"
                            value={data.exit_reason || "-"}
                          />
                          <Attribute
                            name="Autoscroll"
                            value={<AutoscrollSwitch />}
                          />
                        </div>
                      </Card>
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
                            data.engine === ExecutionEngine.EKS ? (
                              <LogRequesterS3
                                runID={data.run_id}
                                status={data.status}
                              />
                            ) : (
                              <LogRequesterCloudWatchLogs
                                runID={data.run_id}
                                status={data.status}
                              />
                            )
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
                              hasLogs={this.props.hasLogs}
                            />
                          }
                          disabled={data.engine !== ExecutionEngine.EKS}
                        />
                        <Tab
                          id={RunTabId.CLOUDTRAIL}
                          title={
                            data.engine !== ExecutionEngine.EKS ? (
                              <Tooltip content="Cloudtrail records are only available for tasks run on EKS.">
                                Cloudtrail Records
                              </Tooltip>
                            ) : (
                              `EKS Cloudtrail Records (${
                                hasCloudtrailRecords
                                  ? get(
                                      data,
                                      ["cloudtrail_notifications", "Records"],
                                      []
                                    ).length
                                  : 0
                              })`
                            )
                          }
                          panel={
                            <CloudtrailRecords data={cloudtrailRecords || []} />
                          }
                          disabled={
                            data.engine !== ExecutionEngine.EKS ||
                            hasCloudtrailRecords === false
                          }
                        />
                      </Tabs>
                    </div>
                  </div>
                </>
              )}
            </Toggler>
          )
        }
        return <Callout title="Run not found" intent={Intent.WARNING} />
      case RequestStatus.NOT_READY:
      default:
        return <Spinner />
    }
  }
}

const ReduxConnectedRun = connected(Run)

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
          <>
            <Helmet>
              <title>
                {`${
                  props.data
                    ? EnhancedRunStatusEmojiMap.get(
                        getEnhancedRunStatus(props.data) as EnhancedRunStatus
                      )
                    : ""
                }
                ${match.params.runID}`}
              </title>
            </Helmet>
            <ReduxConnectedRun
              {...props}
              runID={match.params.runID}
              query={query}
              setQuery={setQuery}
            />
          </>
        )}
      </Request>
    )}
  </QueryParams>
)

export default Connected
