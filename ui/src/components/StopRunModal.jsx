import React from 'react'
import { Modal } from './'

export default function StopRunModal(props) {
  const {
    closeModal,
    stopRun,
  } = props
  return (
    <Modal header="Stop Run" closeModal={closeModal}>
      <div style={{ marginBottom: 12 }}>Are you sure you want to stop this run?</div>
      <button className="button button-error full-width" onClick={stopRun}>Stop</button>
    </Modal>
  )
}
