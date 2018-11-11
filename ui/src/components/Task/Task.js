import React, { Component } from "react"
import PropTypes from "prop-types"
import { isEqual } from "lodash"
import * as requestStateTypes from "../../constants/requestStateTypes"
import api from "../../api"
import TaskContext from "./TaskContext"

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
    const { children } = this.props

    return (
      <TaskContext.Provider value={this.getCtx()}>
        {children}
      </TaskContext.Provider>
    )
  }
}

Task.propTypes = {
  children: PropTypes.node,
  definitionID: PropTypes.string.isRequired,
}

export default Task
