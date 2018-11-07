import React from "react"
import PropTypes from "prop-types"

export default function ModalContainer({ modal }) {
  return (
    <div className="pl-modal-container">
      <div className="pl-modal-overlay" />
      {modal}
    </div>
  )
}

ModalContainer.propTypes = {
  modal: PropTypes.node,
}
ModalContainer.displayName = "ModalContainer"
