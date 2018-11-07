import React from "react"
import PropTypes from "prop-types"

const Topbar = ({ children }) => {
  return (
    <div className="pl-topbar-container">
      <div className="pl-topbar-inner">{children}</div>
    </div>
  )
}

Topbar.displayName = "Topbar"
Topbar.propTypes = {
  children: PropTypes.node,
}

export default Topbar
