import React from "react"
import PropTypes from "prop-types"

const ViewHeader = ({ children, title, actions }) => {
  return (
    <div className="pl-view-header-container">
      <div className="pl-view-header-inner">
        {!!title && <h3>{title}</h3>}
        {!!children && children}
        {!!actions && actions}
      </div>
    </div>
  )
}

ViewHeader.displayName = "ViewHeader"
ViewHeader.propTypes = {
  children: PropTypes.node,
  title: PropTypes.node,
  actions: PropTypes.node,
}

export default ViewHeader
