import * as React from "react"
import { Switch, Route, RouteComponentProps } from "react-router-dom"
import { get } from "lodash"
import Task from "./Task"
import TaskDefinition from "./TaskDefinition"
import RunForm from "../RunForm/RunForm"
import { UpdateTaskForm, CloneTaskForm } from "../TaskForm/TaskForm"

interface ITaskRouterProps extends RouteComponentProps<any> {
  shouldRequestByAlias?: boolean
}

class TaskRouter extends React.PureComponent<ITaskRouterProps> {
  render() {
    const shouldRequestByAlias = this.props.shouldRequestByAlias === true
    const definitionID = get(
      this.props,
      ["match", "params", "definitionID"],
      ""
    )
    const rootPath = get(this.props, ["match", "url"], "")
    const alias = get(this.props, ["match", "params", "alias"], "")

    return (
      <Task
        definitionID={definitionID}
        shouldRequestByAlias={shouldRequestByAlias}
        alias={alias}
      >
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
