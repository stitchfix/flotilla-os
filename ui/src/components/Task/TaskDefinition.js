import React from "react"
import { get, omit } from "lodash"
import TaskContext from "./TaskContext"
import * as requestStateTypes from "../../constants/requestStateTypes"
import View from "../View"
import ViewHeader from "../ViewHeader"
import Loader from "../Loader"
import TaskHistoryTable from "./TaskHistoryTable"

const TaskDefinition = props => {
  return (
    <TaskContext.Consumer>
      {ctx => {
        let title = <Loader mini />
        let actions
        let sidebar = <Loader />

        switch (ctx.requestState) {
          case requestStateTypes.READY:
            title = get(ctx, ["data", "alias"], "")
            sidebar = <div>asdfasdf</div>
            break
          case requestStateTypes.ERROR:
            title = "Error"
            sidebar = "blork"
            break
          case requestStateTypes.NOT_READY:
          default:
            title = "loading"
            sidebar = "loading"
            break
        }

        return (
          <View>
            <ViewHeader title={title} actions={actions} />
            <div>
              {sidebar}
              <TaskHistoryTable definitionID={ctx.definitionID} />
            </div>
          </View>
        )
      }}
    </TaskContext.Consumer>
  )
}

export default props => {
  return (
    <TaskDefinition
      {...omit(props, ["history", "location", "match", "staticContext"])}
    />
  )
}
