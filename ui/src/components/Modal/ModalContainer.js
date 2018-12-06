import React, { Component } from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import { Z_INDICES } from "../../helpers/styles"
import colors from "../../helpers/colors"
import ModalContext from "./ModalContext"

const StyledModalContainer = styled.div`
  width: 100vw;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: flex-start;
  overflow: scroll;
  z-index: ${Z_INDICES.MODAL_CONTAINER};
`

const ModalOverlay = styled.div`
  width: 100vw;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: ${colors.black[0]};
  z-index: ${Z_INDICES.MODAL_OVERLAY};
`

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
          <StyledModalContainer>
            <ModalOverlay />
            {modal}
          </StyledModalContainer>
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
