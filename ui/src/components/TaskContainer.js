import React, { Component } from "react"
import { Route, Switch } from "react-router-dom"
import { connect } from "react-redux"
import { withStateFetch } from "aa-ui-components"
import config from "../config"
import RunForm from "./RunForm"
import TaskDefinitionView from "./TaskDefinitionView"
import CopyTaskForm from "./CopyTaskForm"
import EditTaskForm from "./EditTaskForm"

export class TaskContainer extends Component {
  constructor(props) {
    super(props)
    this.fetch = this.fetch.bind(this)
  }
  componentDidMount() {
    const id = this.props.match.params.definitionId
    this.fetch(id)
  }
  componentWillReceiveProps(nextProps) {
    if (
      this.props.match.params.definitionId !==
      nextProps.match.params.definitionId
    ) {
      this.fetch(nextProps.match.params.definitionId)
    }
  }
  fetch(definitionId) {
    this.props.fetch(`${config.FLOTILLA_API}/task/${definitionId}`)
  }
  render() {
    const { isLoading, data, error, match, dispatch } = this.props
    const rootPath = match.url

    return (
      <Switch>
        <Route
          exact
          path={rootPath}
          render={() => (
            <TaskDefinitionView
              definitionId={match.params.definitionId}
              isLoading={isLoading}
              data={data}
              error={error}
              dispatch={dispatch}
            />
          )}
        />
        <Route
          exact
          path={`${rootPath}/copy`}
          render={() => (
            <CopyTaskForm
              definitionId={match.params.definitionId}
              isLoading={isLoading}
              data={data}
              error={error}
            />
          )}
        />
        <Route
          exact
          path={`${rootPath}/edit`}
          render={() => (
            <EditTaskForm
              definitionId={match.params.definitionId}
              isLoading={isLoading}
              data={data}
              error={error}
              fetch={this.fetch}
            />
          )}
        />
        <Route
          exact
          path={`${rootPath}/run`}
          render={() => (
            <RunForm
              definitionId={match.params.definitionId}
              isLoading={isLoading}
              data={data}
              error={error}
            />
          )}
        />
      </Switch>
    )
  }
}

export default connect()(withStateFetch(TaskContainer))
