import * as React from "react"
import { BrowserRouter, Route, Switch, Redirect } from "react-router-dom"
import Tasks from "./Tasks"
import Task from "./Task"
import CreateTaskForm from "./CreateTaskForm"
import Run from "./Run"
import Runs from "./Runs"
import Navigation from "./Navigation"

const App: React.FunctionComponent = () => (
  <div className="flotilla-app-container bp3-dark">
    <BrowserRouter>
      <Navigation />
      <Switch>
        <Route exact path="/tasks" component={Tasks} />
        <Route exact path="/tasks/create" component={CreateTaskForm} />
        <Route path="/tasks/:definitionID" component={Task} />
        <Route path="/tasks/alias/:alias" component={Task} />
        <Route exact path="/runs" component={Runs} />
        <Route path="/runs/:runID" component={Run} />
        <Redirect from="/" to="/tasks" />
      </Switch>
    </BrowserRouter>
  </div>
)

export default App
