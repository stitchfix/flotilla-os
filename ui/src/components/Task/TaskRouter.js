import React from "react"
import PropTypes from "prop-types"
import { Switch, Route } from "react-router-dom"
import { get, omit } from "lodash"
import Task from "./Task"
import TaskDefinition from "./TaskDefinition"
import RunForm from "../RunForm/RunForm"
import { UpdateTaskForm, CloneTaskForm } from "../TaskForm/TaskForm"

const TaskRouter = ({ rootPath, definitionID }) => (
  <Task definitionID={definitionID}>
    <Switch>
      <Route exact path={rootPath} component={TaskDefinition} />
      <Route exact path={`${rootPath}/run`} component={RunForm} />
      <Route exact path={`${rootPath}/copy`} component={CloneTaskForm} />
      <Route exact path={`${rootPath}/edit`} component={UpdateTaskForm} />
    </Switch>
  </Task>
)

TaskRouter.propTypes = {
  definitionID: PropTypes.string.isRequired,
  rootPath: PropTypes.string.isRequired,
}

export default props => (
  <TaskRouter
    {...omit(props, ["history", "location", "match", "staticContext"])}
    definitionID={get(props, ["match", "params", "definitionID"], "")}
    rootPath={get(props, ["match", "url"], "")}
  />
)
