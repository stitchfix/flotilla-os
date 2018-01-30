import React from "react"
import PropTypes from "prop-types"
import { has } from "lodash"
import { Loader } from "aa-ui-components"

const EmptyTable = props => {
  const loaderContainerStyle = { height: 960 }

  if (props.isLoading) {
    return <Loader containerStyle={loaderContainerStyle} />
  }

  return (
    <div className={`flot-empty-table ${props.error ? "error" : ""}`}>
      {has(props, "title") && (
        <div className="flot-empty-table-title">{props.title}</div>
      )}
      {has(props, "actions") && (
        <div className="flot-empty-table-actions">{props.actions}</div>
      )}
    </div>
  )
}

EmptyTable.propTypes = {
  error: PropTypes.bool.isRequired,
  title: PropTypes.node,
  actions: PropTypes.node,
  isLoading: PropTypes.bool.isRequired,
}

EmptyTable.defaultProps = {
  isLoading: false,
  error: false,
}

EmptyTable.displayName = "EmptyTable"

export default EmptyTable
