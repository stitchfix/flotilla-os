import React from "react"
import RunContext from "./RunContext"
import RunSidebar from "./RunSidebar"
import View from "../View"
import ViewHeader from "../ViewHeader"

const RunView = props => {
  return (
    <RunContext.Consumer>
      {ctx => {
        return (
          <View>
            <ViewHeader title="hi" />
            <div className="flot-detail-view flot-run-view">
              <RunSidebar />
            </div>
          </View>
        )
      }}
    </RunContext.Consumer>
  )
}

export default RunView
