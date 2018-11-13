import React from "react"
import { get } from "lodash"
import RunContext from "./RunContext"
import RunSidebar from "./RunSidebar"
import LogRequester from "./LogRequester"
import View from "../styled/View"

const RunView = props => {
  return (
    <RunContext.Consumer>
      {ctx => {
        return (
          <View>
            <div className="flot-detail-view flot-run-view">
              {/* <RunSidebar /> */}
              <LogRequester
                runID={ctx.runID}
                status={get(ctx, ["data", "status"])}
              />
            </div>
          </View>
        )
      }}
    </RunContext.Consumer>
  )
}

export default RunView
