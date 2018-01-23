import React, { Component } from 'react'

export default class Modal extends Component {
  render() {
    const {
      header,
      children,
      closeModal
    } = this.props
    return (
      <div className="modal">
        <div className="section-container">
          <div className="section-header">
            <div className="section-header-text">{header}</div>
            <button className="button" onClick={closeModal}>Close</button>
          </div>
          <div className="section-content">
            {children}
          </div>
        </div>
      </div>
    )
  }
}
