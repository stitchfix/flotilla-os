import React from "react"
import { connect } from "react-redux"
import Helmet from "react-helmet"
import { Link } from "react-router-dom"
import {
  Button,
  View,
  ViewHeader,
  modalActions,
  intentTypes,
} from "aa-ui-components"
import { get } from "lodash"
import qs from "query-string"
import {
  runStatusTypes,
  invalidRunEnv,
  envNameValueDelimiterChar,
} from "../constants"
import { getRetryEnv, getHelmetTitle } from "../utils/"
import StopRunModal from "./StopRunModal"
import EnhancedRunStatus, { getEnhancedStatus } from "./EnhancedRunStatus"
import RunInfo from "./RunInfo"
import RunLogs from "./RunLogs"

const getHelmetEmoji = enhancedStatus => {
  switch (enhancedStatus) {
    case runStatusTypes.success:
      return "✅"
    case runStatusTypes.failed:
      return "❌"
    default:
      return "⏳"
  }
}

export const RunView = props => {
  const retryEnv = getRetryEnv(get(props.data, "env", []))
  const taskStr = get(props.data, "alias", "")
  const enhancedStatus = getEnhancedStatus(
    get(props.data, "status"),
    get(props.data, "exit_code")
  )
  const helmetEmoji = getHelmetEmoji(enhancedStatus)

  return (
    <div className="pl-view-container">
      <Helmet>
        <title>
          {getHelmetTitle(
            `${helmetEmoji} | Running ${taskStr} (${props.runId})`
          )}
        </title>
      </Helmet>
      <ViewHeader
        title={
          <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
            <div>{props.runId}</div>
            <EnhancedRunStatus
              status={get(props.data, "status")}
              exitCode={get(props.data, "exit_code")}
              iconOnly
            />
          </div>
        }
        actions={
          <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
            <Link
              to={{
                pathname: `/tasks/${get(props.data, "definition_id", "")}/run`,
                search: `?cluster=${get(props.data, "cluster")}&${qs.stringify({
                  env: retryEnv,
                })}`,
              }}
              className="pl-button"
            >
              Retry
            </Link>
            {get(props.data, "status", runStatusTypes.stopped) !==
              runStatusTypes.stopped && (
              <Button
                intent={intentTypes.error}
                onClick={() => {
                  props.dispatch(
                    modalActions.renderModal(
                      <StopRunModal
                        runId={props.runId}
                        definitionId={get(props.data, "definition_id", "")}
                      />
                    )
                  )
                }}
              >
                Stop
              </Button>
            )}
          </div>
        }
      />
      <div className="pl-view-inner" style={{ marginBottom: 0 }}>
        <div className="flot-detail-view flot-run-view">
          <RunInfo {...props} />
          <RunLogs
            runId={props.runId}
            status={get(props.data, "status", undefined)}
          />
        </div>
      </div>
    </div>
  )
}

export default connect()(RunView)
