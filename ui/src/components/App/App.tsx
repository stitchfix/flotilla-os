import * as React from "react"
import { BrowserRouter, Switch, Route, Redirect } from "react-router-dom"
import { CreateTaskForm } from "../TaskForm/TaskForm"
import Tasks from "../Tasks/Tasks"
import ActiveRuns from "../ActiveRuns/ActiveRuns"

const App: React.SFC<{}> = () => (
  <BrowserRouter>
    <Switch>
      <Route exact path="/tasks/create" component={CreateTaskForm} />
      <Route exact path="/runs" component={ActiveRuns} />
      <Route exact path="/tasks" component={Tasks} />
    </Switch>
  </BrowserRouter>
)

export default App
