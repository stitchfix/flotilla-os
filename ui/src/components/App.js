import React, { Component } from "react"
import { PropTypes } from "prop-types"
import { HashRouter, Route, Switch, Redirect, Link } from "react-router-dom"
import { connect } from "react-redux"
import { View, Card, ModalContainer } from "aa-ui-components"
import { fetchSelectOpts } from "../actions/"
import FlotillaTopbar from "./FlotillaTopbar"
import Tasks from "./Tasks"
import ActiveRuns from "./ActiveRuns"
import TaskContainer from "./TaskContainer"
import RunContainer from "./RunContainer"
import CreateTaskForm from "./CreateTaskForm"

export const App = props => {
  const {
    modal: { modalVisible, modal },
    popup: { popupVisible, popup },
  } = props

  return (
    <HashRouter>
      <div>
        {!!modalVisible && !!modal && <ModalContainer modal={modal} />}
        {!!popupVisible && !!popup && popup}
        <FlotillaTopbar />
        <Switch>
          <Route exact path="/tasks/create" component={CreateTaskForm} />
          <Route exact path="/runs" component={ActiveRuns} />
          <Route exact path="/tasks" component={Tasks} />
          <Route path="/tasks/:definitionId" component={TaskContainer} />
          <Route path="/runs/:runId" component={RunContainer} />
          {process.env.NODE_ENV !== "test" ? (
            <Redirect from="/" to="/tasks" />
          ) : null}
        </Switch>
      </div>
    </HashRouter>
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
