import * as React from "react"
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
} from "@blueprintjs/core"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import api from "../api"
import { Run as RunShape, RunStatus } from "../types"
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

export type Props = RequestChildProps<RunShape, { runID: string }> & {
  runID: string
}

export class Run extends React.Component<Props> {
  requestIntervalID: number | undefined

  constructor(props: Props) {
    super(props)
    this.request = this.request.bind(this)
  }

  componentDidMount() {
    const { data } = this.props
    if (data && data.status !== RunStatus.STOPPED) this.setRequestInterval()
  }

  componentDidUpdate(prevProps: Props) {
    if (
      prevProps.requestStatus === RequestStatus.NOT_READY &&
      this.props.requestStatus === RequestStatus.READY &&
      this.props.data &&
      this.props.data.status !== RunStatus.STOPPED
    ) {
      this.setRequestInterval()
    }
    if (this.props.data && this.props.data.status === RunStatus.STOPPED) {
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
      return window.innerHeight - 78 - 50 - 24
    }

    return 720
  }

  render() {
    const { data, requestStatus, runID } = this.props

    if (requestStatus === RequestStatus.READY && data) {
      let btn = null

      if (data.status === RunStatus.STOPPED) {
        btn = (
          <Link
            className={Classes.BUTTON}
            to={{
              pathname: `/tasks/${data.definition_id}/execute`,
              state: {
                cluster: data.cluster,
                env: data.env,
              },
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
        <>
          <ViewHeader
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
            <div className="flotilla-sidebar-view-sidebar">
              <Toggler>
                {({ isVisible, toggleVisibility }) => (
                  <Card style={{ marginBottom: 12 }}>
                    <div className="flotilla-card-header-container">
                      <div className="flotilla-card-header">Attributes</div>
                      <ButtonGroup>
                        <Button
                          onClick={toggleVisibility}
                          rightIcon={isVisible ? "minimize" : "maximize"}
                        >
                          {isVisible ? "Hide" : "Show"}
                        </Button>
                      </ButtonGroup>
                    </div>
                    <Collapse isOpen={isVisible}>
                      <div className="flotilla-attributes-container">
                        <Attribute name="Status" value={<RunTag {...data} />} />
                        <Attribute
                          name="Duration"
                          value={
                            data.started_at && (
                              <Duration
                                start={data.started_at}
                                end={data.finished_at}
                              />
                            )
                          }
                        />
                        <Attribute name="Run ID" value={data.run_id} />
                        <Attribute
                          name="Definition ID"
                          value={data.definition_id}
                        />
                        <Attribute name="Cluster" value={data.cluster} />
                        <Attribute name="Exit Code" value={data.exit_code} />
                        <Attribute
                          name="Exit Reason"
                          value={data.exit_reason}
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
                    </Collapse>
                  </Card>
                )}
              </Toggler>
              <Toggler>
                {({ isVisible, toggleVisibility }) => (
                  <Card>
                    <div className="flotilla-card-header-container">
                      <div className="flotilla-card-header">
                        Environment Variables
                      </div>
                      <ButtonGroup>
                        <Button
                          onClick={toggleVisibility}
                          rightIcon={isVisible ? "minimize" : "maximize"}
                        >
                          {isVisible ? "Hide" : "Show"}
                        </Button>
                      </ButtonGroup>
                    </div>
                    <Collapse isOpen={isVisible}>
                      <EnvList env={data.env} />
                    </Collapse>
                  </Card>
                )}
              </Toggler>
            </div>
            <div className="flotilla-sidebar-view-content">
              <LogRequester
                runID={data.run_id}
                status={data.status}
                height={this.getLogsHeight()}
              />
            </div>
          </div>
        </>
      )
    }

    if (requestStatus === RequestStatus.ERROR) return <div>errro</div>
    return <Spinner />
  }
}

const Connected: React.FunctionComponent<
  RouteComponentProps<{ runID: string }>
> = ({ match }) => (
  <Request<RunShape, { runID: string }>
    requestFn={api.getRun}
    initialRequestArgs={{ runID: match.params.runID }}
  >
    {props => <Run {...props} runID={match.params.runID} />}
  </Request>
)

export default Connected
