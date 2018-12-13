import * as React from "react"
import { BrowserRouter, Switch, Route, Redirect } from "react-router-dom"
import { CreateTaskForm } from "../TaskForm/TaskForm"
import ActiveRuns from "../ActiveRuns/ActiveRuns"

const App: React.SFC<{}> = () => (
  <BrowserRouter>
    <Switch>
      <Route exact path="/tasks/create" component={CreateTaskForm} />
      <Route exact path="/runs" component={ActiveRuns} />
    </Switch>
  </BrowserRouter>
)

export default App
