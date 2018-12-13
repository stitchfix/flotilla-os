import * as React from "react"
import { Switch, Route, RouteComponentProps } from "react-router-dom"
import { get } from "lodash"
import Task from "./Task"
import TaskDefinition from "./TaskDefinition"
import RunForm from "../RunForm/RunForm"
import { UpdateTaskForm, CloneTaskForm } from "../TaskForm/TaskForm"

interface ITaskRouterProps
  extends RouteComponentProps<{ definitionID: string }> {}

class TaskRouter extends React.PureComponent<ITaskRouterProps> {
  render() {
    const definitionID = get(
      this.props,
      ["match", "params", "definitionID"],
      ""
    )
    const rootPath = get(this.props, ["match", "url"], "")
    return (
      <Task definitionID={definitionID}>
        <Switch>
          <Route exact path={rootPath} component={TaskDefinition} />
          <Route exact path={`${rootPath}/run`} component={RunForm} />
          <Route exact path={`${rootPath}/copy`} component={CloneTaskForm} />
          <Route exact path={`${rootPath}/edit`} component={UpdateTaskForm} />
        </Switch>
      </Task>
    )
  }
}

export default TaskRouter
