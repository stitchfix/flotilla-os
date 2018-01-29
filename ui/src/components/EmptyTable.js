import React from "react"
import PropTypes from "prop-types"
import { has } from "lodash"

const EmptyTable = props => (
  <div className={`flot-empty-table ${props.error ? "error" : ""}`}>
    {has(props, "title") && (
      <div className="flot-empty-table-title">{props.title}</div>
    )}
    {has(props, "message") && (
      <div className="flot-empty-table-message">{props.message}</div>
    )}
    {has(props, "actions") && (
      <div className="flot-empty-table-actions">{props.actions}</div>
    )}
  </div>
)

EmptyTable.propTypes = {
  error: PropTypes.bool.isRequired,
  title: PropTypes.node,
  message: PropTypes.node,
  actions: PropTypes.node,
}

EmptyTable.defaultProps = {
  error: false,
}

EmptyTable.displayName = "EmptyTable"

export default EmptyTable
