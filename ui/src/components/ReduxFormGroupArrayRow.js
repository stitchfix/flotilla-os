import React from "react"
import PropTypes from "prop-types"
import { X } from "react-feather"

const ReduxFormGroupArrayRow = ({ children, onRemoveClick }) => {
  return (
    <div className="redux-form-group-array-row">
      {children}
      <button type="button" className="pl-button" onClick={onRemoveClick}>
        <X size={14} />
      </button>
    </div>
  )
}

ReduxFormGroupArrayRow.propTypes = {
  children: PropTypes.node,
  onRemoveClick: PropTypes.func.isRequired,
}

export default ReduxFormGroupArrayRow
