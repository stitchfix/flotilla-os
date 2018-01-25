import React from "react"
import { get, isEmpty } from "lodash"
import RunStatusBar from "./RunStatusBar"

const RunMiniView = ({ data }) => {
  return (
    <RunStatusBar
      startedAt={get(data, "started_at")}
      finishedAt={get(data, "finished_at")}
      status={get(data, "status", "")}
      exitCode={get(data, "exit_code", "")}
    />
  )
}

export default RunMiniView
