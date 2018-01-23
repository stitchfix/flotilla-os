import React from 'react'
import PropTypes from 'prop-types'

export default function ModalContainer({ modal }) {
  return (
    <div className="modal-container">
      <div className="modal-overlay" />
      {modal}
    </div>
  )
}

ModalContainer.propTypes = {
  modal: PropTypes.node,
}
