import React, { Component, createContext } from "react"
import PropTypes from "prop-types"
import { Switch, Route } from "react-router-dom"
import { get, omit, isEqual } from "lodash"

import * as requestStateTypes from "../../constants/requestStateTypes"
import api from "../../api"
import TaskContext from "./TaskContext"
import TaskDefinition from "./TaskDefinition"

class Task extends Component {
  state = {
    inFlight: false,
    error: false,
    data: {},
    requestState: requestStateTypes.NOT_READY,
  }

  componentDidMount() {
    this.requestData()
  }

  componentDidUpdate(prevProps) {
    if (!isEqual(prevProps.definitionID, this.props.definitionID)) {
      this.requestData()
    }
  }

  requestData() {
    this.setState({ inFlight: false, error: false })

    api
      .getTask({ definitionID: this.props.definitionID })
      .then(data => {
        this.setState({
          inFlight: false,
          data,
          error: false,
          requestState: requestStateTypes.READY,
        })
      })
      .catch(error => {
        this.setState({
          inFlight: false,
          error,
          requestState: requestStateTypes.ERROR,
        })
      })
  }

  getCtx() {
    const { definitionID } = this.props
    return {
      ...this.state,
      definitionID,
    }
  }

  render() {
    const { rootPath } = this.props

    return (
      <TaskContext.Provider value={this.getCtx()}>
        <Switch>
          <Route exact path={rootPath} component={TaskDefinition} />
        </Switch>
      </TaskContext.Provider>
    )
  }
}

Task.propTypes = {
  definitionID: PropTypes.string.isRequired,
  rootPath: PropTypes.string.isRequired,
}

export default props => (
  <Task
    {...omit(props, ["history", "location", "match", "staticContext"])}
    definitionID={get(props, ["match", "params", "definitionID"], "")}
    rootPath={get(props, ["match", "url"], "")}
  />
)
