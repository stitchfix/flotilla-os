import React from "react"
import PropTypes from "prop-types"

const Field = ({ label, children, description, error }) => (
  <div className="pl-form-group full-width">
    <label className="pl-form-group-label">{label}</label>
    {children}
    {!!error && <div className="pl-form-group-error">{error}</div>}
    {!!description && (
      <div className="pl-form-group-description">{description}</div>
    )}
  </div>
)

Field.displayName = "Field"

Field.propTypes = {
  children: PropTypes.node.isRequired,
  description: PropTypes.string,
  error: PropTypes.any,
  label: PropTypes.string.isRequired,
}

Field.defaultProps = {
  error: false,
}

export default Field
