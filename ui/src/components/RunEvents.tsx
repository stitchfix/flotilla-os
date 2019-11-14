import * as React from "react"
import { RunStatus } from "../types"
import Request, { RequestStatus } from "./Request"
import api from "../api"
import { ListRunEventsResponse } from "../types"
import ErrorCallout from "./ErrorCallout"
import { Spinner, Callout, Card, Tag } from "@blueprintjs/core"

type Props = {
  runID: string
  status: RunStatus
}

const RunEvents: React.FC<Props> = ({ runID, status }) => (
  <Request<ListRunEventsResponse, string>
    requestFn={api.listRunEvents}
    initialRequestArgs={runID}
  >
    {({ data, requestStatus, isLoading, error }) => {
      switch (requestStatus) {
        case RequestStatus.ERROR:
          return <ErrorCallout error={error} />
        case RequestStatus.READY:
          if (data && data.run_events !== null) {
            return data.run_events.map(evt => (
              <Card style={{ marginBottom: 12 }}>
                <div className="flotilla-card-header-container">
                  <div className="flotilla-card-header">
                    {evt.timestamp} <Tag>{evt.reason}</Tag>
                  </div>
                </div>
                {evt.message}
              </Card>
            ))
          }
          return <Callout>No events found.</Callout>
        case RequestStatus.NOT_READY:
        default:
          return <Spinner />
      }
    }}
  </Request>
)

export default RunEvents
