import React, { Component } from 'react'
import { connect } from 'react-redux'
import { TaskForm, TaskFormNav } from '../'
import { submitTask } from '../../actions/'
import { TaskFormTypes } from '../../constants/'

class TaskFormContainer extends Component {
  handleCancel() {
    const { formType, taskID, router } = this.props
    const pathname = formType === TaskFormTypes.new ?
      `/tasks` :
      `/tasks/${taskID}`
    router.push({ pathname })
  }
  handleSave(values) {
    const { formType, dispatch, taskID } = this.props

    dispatch(submitTask({
      values,
      formType,
      id: taskID
    }))
  }
  render() {
    const { formType } = this.props

    return (
      <div className="view-container">
        <TaskFormNav
          formType={formType}
          onCancel={() => { this.handleCancel() }}
        />
        <div className="view">
          <div className="layout-standard flex ff-cn j-fs a-c">
            <TaskForm
              formType={formType}
              onSubmit={(values) => { this.handleSave(values) }}
            />
          </div>
        </div>
      </div>
    )
  }
}

function mapStateToProps(state, ownProps) {
  let formType
  let taskID
  let taskAlias

  // Determine form type based on location.
  if (ownProps.location.pathname.endsWith('/create-task')) {
    formType = TaskFormTypes.new
  } else if (!!ownProps.params.taskID) {
    taskAlias = state.task.task.alias
    taskID = ownProps.params.taskID
    if (ownProps.location.pathname.endsWith('copy')) {
      formType = TaskFormTypes.copy
    } else if (ownProps.location.pathname.endsWith('edit')) {
      formType = TaskFormTypes.edit
    }
  }

  return ({ formType, taskID, taskAlias })
}

export default connect(mapStateToProps)(TaskFormContainer)
