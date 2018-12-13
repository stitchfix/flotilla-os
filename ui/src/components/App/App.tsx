import * as React from "react"
import { BrowserRouter, Switch, Route, Redirect } from "react-router-dom"
import { CreateTaskForm } from "../TaskForm/TaskForm"
import Tasks from "../Tasks/Tasks"
import ActiveRuns from "../ActiveRuns/ActiveRuns"
import TaskRouter from "../Task/TaskRouter"
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
          <Route path="/tasks/:definitionID" component={TaskRouter} />
        </Switch>
      </PopupContainer>
    </ModalContainer>
  </BrowserRouter>
)

export default App
