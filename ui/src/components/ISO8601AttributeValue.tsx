import * as React from "react"
import moment from "moment"
import { Classes } from "@blueprintjs/core"

const ISO8601AttributeValue: React.FunctionComponent<{
  time: string | null | undefined
  inline?: boolean
  verbose?: boolean
}> = ({ time, inline, verbose }) => {
  return (
    <div
      style={{
        display: "flex",
        flexDirection: inline && inline === true ? "row" : "column",
        alignItems: inline && inline === true ? "flex-end" : "flex-start",
      }}
    >
      <div style={{ marginRight: inline && inline === true ? 4 : 0 }}>
        {time !== null && time !== undefined ? moment(time).fromNow() : "-"}
      </div>
      {verbose && time !== null && time !== undefined && (
        <div className={Classes.TEXT_SMALL}>{time.substr(0, 19)}</div>
      )}
    </div>
  )
}

ISO8601AttributeValue.defaultProps = {
  verbose: true,
}

export default ISO8601AttributeValue
