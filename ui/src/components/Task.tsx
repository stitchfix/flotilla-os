import * as React from "react"
import { Switch, Route, RouteComponentProps } from "react-router-dom"
import { get } from "lodash"
import Request, { ChildProps, RequestStatus } from "./Request"
import api from "../api"
import { Task as TaskShape, Task as TaskTypeDef } from "../types"
import TaskDetails from "./TaskDetails"
import UpdateTaskForm from "./UpdateTaskForm"
import RunForm from "./RunForm"
import CreateTaskForm from "./CreateTaskForm"
import ErrorCallout from "./ErrorCallout"
import { Spinner } from "@blueprintjs/core"

export type TaskCtx = ChildProps<TaskShape, { definitionID: string }> & {
  basePath: string
  definitionID: string
}

export const TaskContext = React.createContext<TaskCtx>({
  data: null,
  requestStatus: RequestStatus.NOT_READY,
  isLoading: false,
  error: null,
  request: () => {},
  basePath: "", // TODO: maybe this is not required.
  definitionID: "",
})

export const Task: React.FunctionComponent<TaskCtx> = props => {
  return (
    <TaskContext.Provider value={props}>
      <Switch>
        <Route exact path={props.basePath} component={TaskDetails} />
        <Route
          exact
          path={`${props.basePath}/update`}
          component={UpdateTaskForm}
        />
        <Route
          exact
          path={`${props.basePath}/copy`}
          render={routerProps => (
            <TaskContext.Consumer>
              {ctx => {
                switch (ctx.requestStatus) {
                  case RequestStatus.ERROR:
                    return <ErrorCallout error={ctx.error} />
                  case RequestStatus.READY:
                    return (
                      <CreateTaskForm
                        {...routerProps}
                        onSuccess={(data: TaskTypeDef) => {
                          ctx.request({ definitionID: data.definition_id })
                        }}
                        initialValues={{
                          env: get(props, ["data", "env"], []),
                          image: get(props, ["data", "image"], ""),
                          group_name: get(props, ["data", "group_name"], ""),
                          memory: get(props, ["data", "memory"], ""),
                          command: get(props, ["data", "command"], ""),
                          tags: get(props, ["data", "tags"], []),
                          alias: "",
                        }}
                      />
                    )
                  case RequestStatus.NOT_READY:
                    return <Spinner />
                  default:
                    return null
                }
              }}
            </TaskContext.Consumer>
          )}
        />
        <Route exact path={`${props.basePath}/execute`} component={RunForm} />
      </Switch>
    </TaskContext.Provider>
  )
}

type ConnectedProps = RouteComponentProps<{ definitionID: string }>
const Connected: React.FunctionComponent<ConnectedProps> = ({ match }) => (
  <Request<TaskShape, { definitionID: string }>
    requestFn={api.getTask}
    initialRequestArgs={{ definitionID: match.params.definitionID }}
  >
    {props => (
      <Task
        {...props}
        basePath={match.path}
        definitionID={match.params.definitionID}
      />
    )}
  </Request>
)

export default Connected
