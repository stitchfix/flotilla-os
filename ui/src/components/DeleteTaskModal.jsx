import React from 'react'
import { connect } from 'react-redux'
import { Modal, Loader } from './'

function DeleteTaskModal(props) {
  const {
    closeModal,
    deleteTask,
    isDeleting,
  } = props
  return (
    <Modal header="Delete Task" closeModal={closeModal}>
      <div style={{ marginBottom: 12 }}>Are you sure you want to delete this task?</div>
      <button className="button button-error full-width" onClick={deleteTask}>
        {isDeleting ? <Loader mini /> : 'Delete Task'}
      </button>
    </Modal>
  )
}

const mapStateToProps = state => ({
  isDeleting: state.task._isDeleting
})

export default connect(mapStateToProps)(DeleteTaskModal)
