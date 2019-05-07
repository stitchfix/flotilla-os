import * as React from "react"
import {
  BrowserRouter,
  Switch,
  Route,
  Redirect,
  RouteComponentProps,
} from "react-router-dom"
import CreateTaskForm from "../TaskForm/CreateTaskForm"
import Tasks from "../Tasks/Tasks"
import ActiveRuns from "../Runs/Runs"
import TaskRouter from "../Task/TaskRouter"
import Run from "../Run/Run"
import ModalContainer from "../Modal/ModalContainer"
import PopupContainer from "../Popup/PopupContainer"

const App: React.SFC<{}> = () => (
  <BrowserRouter>
    <ModalContainer>
      <PopupContainer>
        <Switch>
          <Route exact path="/tasks/create" component={CreateTaskForm} />
          <Route exact path="/runs" component={ActiveRuns} />
          <Route exact path="/tasks" component={Tasks} />
          <Route
            path="/tasks/alias/:alias"
            component={(props: RouteComponentProps<any>) => (
              <TaskRouter {...props} shouldRequestByAlias />
            )}
          />
          <Route path="/tasks/:definitionID" component={TaskRouter} />
          <Route path="/runs/:runID" component={Run} />
          {process.env.NODE_ENV !== "test" ? (
            <Redirect from="/" to="/tasks" />
          ) : null}
        </Switch>
      </PopupContainer>
    </ModalContainer>
  </BrowserRouter>
)

export default App
