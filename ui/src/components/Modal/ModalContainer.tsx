import * as React from "react"
import styled from "styled-components"
import { Z_INDICES } from "../../helpers/styles"
import colors from "../../helpers/colors"
import ModalContext from "./ModalContext"
import { IModalContext } from "../../.."

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

interface IModalContainerState {
  isVisible: boolean
  modal: React.ReactNode | undefined
}

class ModalContainer extends React.Component<{}, IModalContainerState> {
  static displayName = "ModalContainer"
  state = {
    isVisible: false,
    modal: undefined,
  }

  renderModal = (modal: React.ReactNode): void => {
    this.setState({ isVisible: true, modal })
  }

  unrenderModal = (): void => {
    this.setState({ isVisible: false, modal: undefined })
  }

  getCtx(): IModalContext {
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

export default ModalContainer
