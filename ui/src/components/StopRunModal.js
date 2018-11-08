import React from "react"
import PropTypes from "prop-types"
import ConfirmModal from "./ConfirmModal"
import api from "../api"

const StopRunModal = props => (
  <ConfirmModal
    body="Are you sure you want to stop this run?"
    requestFn={api.stopRun}
    getRequestArgs={() => {
      return {
        definitionID: props.definitionID,
        runID: props.runID,
      }
    }}
    onSuccess={() => {
      props.history.push("/tasks")
    }}
  />
)

StopRunModal.displayName = "StopRunModal"

StopRunModal.propTypes = {
  definitionID: PropTypes.string.isRequired,
  history: PropTypes.shape({
    push: PropTypes.func,
  }),
  runID: PropTypes.string.isRequired,
}

StopRunModal.defaultProps = {}

export default StopRunModal
