import React from "react"
import PropTypes from "prop-types"

const Modal = ({ children }) => {
  return <div className="pl-modal">{children}</div>
}

Modal.displayName = "Modal"
Modal.propTypes = {
  children: PropTypes.node,
}

export default Modal
