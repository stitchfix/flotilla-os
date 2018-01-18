import React, { Component } from "react"
import { PropTypes } from "prop-types"
import {
  HashRouter,
  Route,
  Switch,
  NavLink,
  Redirect,
  Link,
} from "react-router-dom"
import { connect } from "react-redux"
// import ReactCSSTransitionGroup from "react-addons-css-transition-group"
import { View, Card, Topbar, ModalContainer } from "aa-ui-components"
import { fetchSelectOpts } from "../actions/"
import Tasks from "./Tasks"
import ActiveRuns from "./ActiveRuns"
import TaskContainer from "./TaskContainer"
import RunContainer from "./RunContainer"
import CreateTaskForm from "./CreateTaskForm"

const containerStyle = {
  background: `rgba(0, 0, 0, 0.28)`,
  border: "none",
  borderRadius: 8,
}
const contentStyle = {
  padding: 16,
}
const filterBarStyle = {
  marginTop: 12,
  marginBottom: 12,
  paddingTop: 12,
  paddingBottom: 12,
  borderTop: "1px solid rgba(255, 255, 255, 0.2)",
  borderBottom: "1px solid rgba(255, 255, 255, 0.2)",
}

const FilterBar = props => (
  <div
    className="flex ff-rn j-fs a-c with-horizontal-child-margin"
    style={filterBarStyle}
  >
    <h4>FILTER BY</h4>
    {props.filters.map((filter, i) => <div key={i}>{filter}</div>)}
  </div>
)

const Test = props => {
  return (
    <View noHeader>
      <Card containerStyle={containerStyle}>
        <div className="flex ff-cn j-fs a-s full-width">
          <h1>Tasks</h1>
          <FilterBar filters={["Alias", "Group", "Image"]} />
        </div>
      </Card>
    </View>
  )
}

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
        <Topbar>
          <div className="pl-topbar-section">
            <div className="pl-topbar-app-name">
              <Link to="/">
                <img
                  src="favicon.png"
                  alt="stitchfix-logo"
                  style={{
                    width: 32,
                    height: 32,
                    borderRadius: 6,
                    marginRight: 6,
                    transform: "translateY(4px)",
                  }}
                />
              </Link>
            </div>
            <NavLink className="pl-topbar-nav-link" to="/tasks">
              Tasks
            </NavLink>
            <NavLink className="pl-topbar-nav-link" to="/runs">
              Runs
            </NavLink>
          </div>
        </Topbar>
        <Switch>
          <Route exact path="/tasks/create" component={CreateTaskForm} />
          <Route exact path="/runs" component={ActiveRuns} />
          <Route exact path="/tasks" component={Tasks} />
          <Route path="/tasks/:definitionId" component={TaskContainer} />
          <Route path="/runs/:runId" component={RunContainer} />
          <Route path="/weee" component={Test} />
          {process.env.NODE_ENV === "production" ? (
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
