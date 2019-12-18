import * as React from "react"
import { RunStatus, RunTabId } from "../types"
import Request, { RequestStatus } from "./Request"
import api from "../api"
import { ListRunEventsResponse } from "../types"
import ErrorCallout from "./ErrorCallout"
import { Spinner, Callout, Card, Tag, Button, Intent } from "@blueprintjs/core"
import QueryParams from "./QueryParams"
import { RUN_TAB_ID_QUERY_KEY } from "../constants"

type Props = {
  runID: string
  status: RunStatus
  hasLogs: boolean
}

const RunEvents: React.FC<Props> = ({ runID, status, hasLogs }) => (
  <QueryParams>
    {({ setQuery }) => (
      <Request<ListRunEventsResponse, string>
        requestFn={api.listRunEvents}
        initialRequestArgs={runID}
      >
        {({ data, requestStatus, isLoading, error }) => {
          switch (requestStatus) {
            case RequestStatus.ERROR:
              return <ErrorCallout error={error} />
            case RequestStatus.READY:
              let viewLogsCallout = (
                <Callout
                  intent={Intent.PRIMARY}
                  title="Logs Available!"
                  style={{ marginTop: 24 }}
                >
                  <Button
                    intent={Intent.PRIMARY}
                    onClick={() => {
                      setQuery({ [RUN_TAB_ID_QUERY_KEY]: RunTabId.LOGS })
                    }}
                  >
                    View Logs
                  </Button>
                </Callout>
              )
              if (data && data.pod_events !== null) {
                return (
                  <>
                    <div>
                      {data.pod_events.map((evt, i) => (
                        <Card style={{ marginBottom: 12 }} key={i}>
                          <div className="flotilla-card-header-container">
                            <div className="flotilla-card-header">
                              {evt.timestamp} <Tag>{evt.reason}</Tag>
                            </div>
                          </div>
                          {evt.message}
                        </Card>
                      ))}
                    </div>
                    {hasLogs && viewLogsCallout}
                  </>
                )
              }
              return (
                <>
                  <Callout>No events found.</Callout>
                  {hasLogs && viewLogsCallout}
                </>
              )
            case RequestStatus.NOT_READY:
            default:
              return <Spinner />
          }
        }}
      </Request>
    )}
  </QueryParams>
)

export default RunEvents
