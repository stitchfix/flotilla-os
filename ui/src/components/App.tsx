import * as React from "react"
import { BrowserRouter, Route, Switch, Redirect } from "react-router-dom"
import Tasks from "./Tasks"
import Task from "./Task"
import CreateTaskForm from "./CreateTaskForm"
import Run from "./Run"
import Runs from "./Runs"
import Navigation from "./Navigation"
import ls from "../localstorage"
import { LOCAL_STORAGE_IS_ONBOARDED_KEY } from "../constants"
import Toaster from "./Toaster"
import { Intent } from "@blueprintjs/core"
import { connect, ConnectedProps } from "react-redux"
import { RootState } from "../state/store"
import { toggleDialogVisibilityChange } from "../state/settings"

const connector = connect()

class App extends React.Component<ConnectedProps<typeof connector>> {
  componentDidMount() {
    this.checkOnboardingStatus()
  }

  checkOnboardingStatus() {
    ls.getItem<boolean>(LOCAL_STORAGE_IS_ONBOARDED_KEY).then(res => {
      if (res !== true) {
        Toaster.show({
          icon: "clean",
          message:
            "You can now configure your global settings like default owner ID via the Settings menu.",
          timeout: 0,
          intent: Intent.PRIMARY,
          action: {
            onClick: () => {
              ls.setItem<boolean>(LOCAL_STORAGE_IS_ONBOARDED_KEY, true).then(
                () => {
                  this.props.dispatch(toggleDialogVisibilityChange(true))
                }
              )
            },
            text: "Open settings menu",
          },
          onDismiss: () => {
            ls.setItem<boolean>(LOCAL_STORAGE_IS_ONBOARDED_KEY, true)
          },
        })
      }
    })
  }

  render() {
    return (
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
  }
}

export default connector(App)
