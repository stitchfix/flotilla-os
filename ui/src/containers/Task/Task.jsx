import React, { Component } from 'react'
import { connect } from 'react-redux'
import { isEmpty } from 'lodash'
import TaskNav from './TaskNav'
import TaskHistory from './TaskHistory'
import TaskInfo from './TaskInfo'
import { deleteTask, renderModal, unrenderModal } from '../../actions/'
import { DeleteTaskModal, ErrorView, Loader } from '../../components/'

class Task extends Component {
  constructor(props) {
    super(props)
    this.renderDeleteTaskModal = this.renderDeleteTaskModal.bind(this)
  }
  renderDeleteTaskModal() {
    const { dispatch, task } = this.props

    const modal = (
      <DeleteTaskModal
        deleteTask={() => {
          dispatch(deleteTask({
            taskID: task.definition_id
          }, () => {
            dispatch(unrenderModal())
            this.props.router.push(`/tasks`)
          }))
        }}
        closeModal={() => { dispatch(unrenderModal()) }}
      />
    )
    dispatch(renderModal({ modal }))
  }
  render() {
    const { hasError, task, error } = this.props

    let viewContent

    if (!!hasError && !!error) {
      viewContent = <ErrorView error={error} />
    } else if (!isEmpty(task)) {
      viewContent = (
        <div className="layout-detail sidebar-left">
          <div className="layout-detail-sidebar">
            <TaskInfo />
          </div>
          <div className="layout-detail-content">
            <TaskHistory />
          </div>
        </div>
      )
    } else {
      viewContent = <Loader />
    }

    return (
      <div className="view-container">
        <TaskNav
          onDeleteButtonClick={this.renderDeleteTaskModal}
          hasError={hasError}
        />
        <div className="view">
          {viewContent}
        </div>
      </div>
    )
  }
}

const mapStateToProps = state => ({
  task: state.task.task,
  isFetching: state.task.isFetching,
  hasError: state.task.hasError,
  error: state.task.error,
})

export default connect(mapStateToProps)(Task)
