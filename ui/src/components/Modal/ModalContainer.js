import React, { Component } from "react"
import PropTypes from "prop-types"
import ModalContext from "./ModalContext"

class ModalContainer extends Component {
  state = {
    isVisible: false,
    modal: null,
  }

  renderModal = modal => {
    this.setState({ isVisible: true, modal })
  }

  unrenderModal = () => {
    this.setState({ isVisible: false, modal: null })
  }

  getCtx() {
    return {
      renderModal: this.renderModal,
      unrenderModal: this.unrenderModal,
    }
  }

  render() {
    const { modal, isVisible } = this.state

    return (
      <ModalContext.Provider value={this.getCtx()}>
        {!!isVisible && (
          <div className="pl-modal-container">
            <div className="pl-modal-overlay" />
            {modal}
          </div>
        )}
        {this.props.children}
      </ModalContext.Provider>
    )
  }
}

ModalContainer.displayName = "ModalContainer"

ModalContainer.propTypes = {
  children: PropTypes.node,
}

export default ModalContainer
