import React, { Component } from 'react'
import { connect } from 'react-redux'
import { isEqual } from 'lodash'
import { fetchTask, resetTask } from '../../actions/'

class TaskContainer extends Component {
  componentDidMount() {
    const { params, location, dispatch } = this.props
    let id
    if (!!params.taskID) {
      id = params.taskID
    } else if (!!location.state.taskID) {
      // Note: this is to handle the case where the route is passed
      // the task's definition ID via state in the <RunRow> component,
      // location at /src/containers/Main/Runs.jsx
      id = location.state.taskID
    }

    dispatch(fetchTask({ id }))
  }
  componentDidUpdate(prevProps) {
    const { params, dispatch } = this.props
    if ((!!params.taskID && !!prevProps.params.taskID) &&
        !isEqual(params.taskID, prevProps.params.taskID)) {
      dispatch(fetchTask({ id: params.taskID }))
    }
  }
  componentWillUnmount() {
    this.props.dispatch(resetTask())
  }
  render() {
    return (
      <div>{this.props.children}</div>
    )
  }
}

export default connect()(TaskContainer)
