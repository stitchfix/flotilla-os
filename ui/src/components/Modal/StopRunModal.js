import React from "react"
import PropTypes from "prop-types"
import ConfirmModal from "./ConfirmModal"
import api from "../../api"

const StopRunModal = props => {
  return (
    <ConfirmModal
      body="Are you sure you want to stop this run?"
      requestFn={api.stopRun}
      getRequestArgs={() => ({
        definitionID: props.definitionID,
        runID: props.runID,
      })}
    />
  )
}

StopRunModal.displayName = "StopRunModal"

StopRunModal.propTypes = {
  definitionID: PropTypes.string.isRequired,
  runID: PropTypes.string.isRequired,
}

StopRunModal.defaultProps = {}

export default StopRunModal
