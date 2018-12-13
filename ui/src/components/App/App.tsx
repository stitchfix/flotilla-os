import * as React from "react"
import { BrowserRouter, Switch, Route, Redirect } from "react-router-dom"
import { CreateTaskForm } from "../TaskForm/TaskForm"

const App: React.SFC<{}> = () => (
  <BrowserRouter>
    <Switch>
      <Route exact path="/" component={CreateTaskForm} />
    </Switch>
  </BrowserRouter>
)

export default App
