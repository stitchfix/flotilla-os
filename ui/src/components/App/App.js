import React from "react"
import { BrowserRouter, Route, Switch, Redirect } from "react-router-dom"
import ActiveRuns from "../ActiveRuns/ActiveRuns"
import Tasks from "../Tasks/Tasks"
import TaskRouter from "../Task/TaskRouter"
import Run from "../Run/Run"
import { CreateTaskForm } from "../TaskForm/TaskForm"
import ModalContainer from "../Modal/ModalContainer"
import PopupContainer from "../Popup/PopupContainer"
import Navigation from "./Navigation"

const App = () => (
  <BrowserRouter>
    <ModalContainer>
      <PopupContainer>
        <Navigation />
        <Switch>
          <Route exact path="/tasks/create" component={CreateTaskForm} />
          <Route exact path="/runs" component={ActiveRuns} />
          <Route exact path="/tasks" component={Tasks} />
          {/* <Route path="/tasks/alias/:alias" component={TaskByAliasRedirect} /> */}
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
