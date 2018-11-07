import React from "react"
import { PropTypes } from "prop-types"
import { BrowserRouter, Route, Switch, Redirect } from "react-router-dom"
import { connect } from "react-redux"

import ActiveRuns from "./ActiveRuns"
import CreateTaskForm from "./CreateTaskForm"
import FlotillaTopbar from "./FlotillaTopbar"
import ModalContainer from "./ModalContainer"
import RunContainer from "./RunContainer"
import TaskByAliasRedirect from "./TaskByAliasRedirect"
import TaskContainer from "./TaskContainer"
import Tasks from "./Tasks"

export const App = props => {
  const {
    modal: { modalVisible, modal },
    popup: { popupVisible, popup },
  } = props

  return (
    <BrowserRouter>
      <div>
        {!!modalVisible && !!modal && <ModalContainer modal={modal} />}
        {!!popupVisible && !!popup && popup}
        <FlotillaTopbar />
        <Switch>
          <Route exact path="/tasks/create" component={CreateTaskForm} />
          <Route exact path="/runs" component={ActiveRuns} />
          <Route exact path="/tasks" component={Tasks} />
          <Route path="/tasks/alias/:alias" component={TaskByAliasRedirect} />
          <Route path="/tasks/:definitionId" component={TaskContainer} />
          <Route path="/runs/:runId" component={RunContainer} />
          {process.env.NODE_ENV !== "test" ? (
            <Redirect from="/" to="/tasks" />
          ) : null}
        </Switch>
      </div>
    </BrowserRouter>
  )
}

App.propTypes = {
  modal: PropTypes.shape({
    modalVisible: PropTypes.bool,
    modal: PropTypes.node,
  }),
  popup: PropTypes.shape({
    popupVisible: PropTypes.bool,
    popup: PropTypes.node,
  }),
}

const mapStateToProps = ({ modal, popup }) => ({ modal, popup })

export default connect(mapStateToProps)(App)
