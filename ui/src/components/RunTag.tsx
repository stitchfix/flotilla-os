import * as React from "react"
import { Run } from "../types"
import { Tag, Colors } from "@blueprintjs/core"
import { RUN_STATUS_COLOR_MAP } from "../constants"
import getEnhancedRunStatus from "../helpers/getEnhancedRunStatus"

const RunTag: React.FunctionComponent<Run> = run => {
  const enhancedStatus = getEnhancedRunStatus(run)

  return (
    <Tag
      style={{
        color: Colors.WHITE,
        fontWeight: 500,
        background: RUN_STATUS_COLOR_MAP.get(enhancedStatus) || "",
      }}
    >
      {enhancedStatus}
    </Tag>
  )
}

export default RunTag
