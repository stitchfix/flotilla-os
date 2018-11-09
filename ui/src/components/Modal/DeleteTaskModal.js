import React from "react"
import PropTypes from "prop-types"
import { withRouter } from "react-router-dom"
import ConfirmModal from "./ConfirmModal"
import api from "../../api"

const DeleteTaskModal = props => (
  <ConfirmModal
    body="Are you sure you want to delete this task?"
    getRequestArgs={() => ({ definitionID: props.definitionID })}
    requestFn={api.deleteTask}
    onSuccess={() => {
      props.history.push("/tasks")
    }}
  />
)

DeleteTaskModal.displayName = "DeleteTaskModal"

DeleteTaskModal.propTypes = {
  definitionID: PropTypes.string.isRequired,
  history: PropTypes.shape({
    push: PropTypes.func,
  }),
}

DeleteTaskModal.defaultProps = {}

export default withRouter(DeleteTaskModal)
