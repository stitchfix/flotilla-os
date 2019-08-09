import * as React from "react"
import { Link, RouteComponentProps } from "react-router-dom"
import {
  Card,
  Spinner,
  Button,
  ButtonGroup,
  Intent,
  Classes,
} from "@blueprintjs/core"
import Request, {
  ChildProps as RequestChildProps,
  RequestStatus,
} from "./Request"
import api from "../api"
import { Run as RunShape, RunStatus } from "../types"
import Attribute from "./Attribute"
import Logs from "./Logs"
import EnvList from "./EnvList"
import ViewHeader from "./ViewHeader"
import StopRunButton from "./StopRunButton"

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
    this.requestIntervalID = window.setInterval(this.request, 5000)
  }

  clearRequestInterval() {
    window.clearInterval(this.requestIntervalID)
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
            Retry
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
              <Card style={{ marginBottom: 12 }}>
                <div className="flotilla-card-header">Attributes</div>
                <div className="flotilla-attributes-container">
                  <Attribute name="Run ID" value={data.run_id} />
                  <Attribute name="Definition ID" value={data.definition_id} />
                  <Attribute name="Cluster" value={data.cluster} />
                  <Attribute name="Status" value={data.status} />
                  <Attribute name="Exit Code" value={data.exit_code} />
                  <Attribute name="Exit Reason" value={data.exit_reason} />
                  <Attribute name="Started At" value={data.started_at} />
                  <Attribute name="Finished At" value={data.finished_at} />
                  <Attribute name="Image" value={data.image} />
                </div>
              </Card>
              <Card>
                <div className="flotilla-card-header">
                  Environment Variables
                </div>
                <EnvList env={data.env} />
              </Card>
            </div>
            <div className="flotilla-sidebar-view-content">
              <Logs
                runID={runID}
                status={data.status}
                requestFn={api.getRunLog}
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
