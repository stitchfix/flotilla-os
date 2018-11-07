import React from "react"
import PropTypes from "prop-types"
import { has } from "lodash"
import Loader from "./Loader"

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
  actions: PropTypes.node,
  error: PropTypes.bool.isRequired,
  isLoading: PropTypes.bool.isRequired,
  title: PropTypes.node,
}

EmptyTable.defaultProps = {
  isLoading: false,
  error: false,
}

EmptyTable.displayName = "EmptyTable"

export default EmptyTable
