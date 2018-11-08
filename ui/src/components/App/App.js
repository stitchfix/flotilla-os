import React from "react"
import { BrowserRouter, Route, Switch, Redirect } from "react-router-dom"

import FlotillaTopbar from "../FlotillaTopbar"
import RunContainer from "../RunContainer"
import TaskByAliasRedirect from "../TaskByAliasRedirect"

import ActiveRuns from "../ActiveRuns/ActiveRuns"
import Tasks from "../Tasks/Tasks"
import Task from "../Task/Task"
import { CreateTaskForm } from "../TaskForm/TaskForm"

import ModalContainer from "./Modal"
import PopupContainer from "./Popup"

const App = () => (
  <BrowserRouter>
    <ModalContainer>
      <PopupContainer>
        <FlotillaTopbar />
        <Switch>
          <Route exact path="/tasks/create" component={CreateTaskForm} />
          <Route exact path="/runs" component={ActiveRuns} />
          <Route exact path="/tasks" component={Tasks} />
          <Route path="/tasks/alias/:alias" component={TaskByAliasRedirect} />
          <Route path="/tasks/:definitionID" component={Task} />
          <Route path="/runs/:runId" component={RunContainer} />
          {process.env.NODE_ENV !== "test" ? (
            <Redirect from="/" to="/tasks" />
          ) : null}
        </Switch>
      </PopupContainer>
    </ModalContainer>
  </BrowserRouter>
)

export default App
