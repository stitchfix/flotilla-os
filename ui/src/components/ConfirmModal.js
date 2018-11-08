import React, { Component } from "react"
import PropTypes from "prop-types"
import Button from "./Button"
import ButtonGroup from "./ButtonGroup"
import Card from "./Card"
import ModalContext from "./Modal/ModalContext"
import Modal from "./Modal/Modal"
import PopupContext from "./Popup/PopupContext"
import Popup from "./Popup/Popup"
import * as intentTypes from "../constants/intentTypes"

class ConfirmModal extends Component {
  state = {
    inFlight: false,
    error: false,
  }

  handleConfirm = () => {
    const {
      requestFn,
      renderPopup,
      unrenderModal,
      unrenderPopup,
      onSuccess,
      onFailure,
      getRequestArgs,
    } = this.props

    this.setState({ inFlight: true, error: false })

    requestFn(getRequestArgs())
      .then(res => {
        renderPopup(
          <Popup
            title="Success!"
            message="Action was completed successfully."
            intent={intentTypes.success}
            hide={unrenderPopup}
          />
        )
        unrenderModal()
        onSuccess(res)
      })
      .catch(error => {
        this.setState({ inFlight: false, error })

        renderPopup(
          <Popup
            title="Error!"
            message="An error occurred."
            intent={intentTypes.error}
            autohide={false}
            hide={unrenderPopup}
          />
        )

        onFailure()
      })
  }

  render() {
    const { unrenderModal, body, title } = this.props
    const { inFlight, error } = this.state

    return (
      <Modal>
        <Card
          header={title}
          footer={
            <ButtonGroup>
              <Button onClick={unrenderModal}>Cancel</Button>
              <Button
                intent={intentTypes.error}
                onClick={this.handleConfirm}
                isLoading={inFlight}
              >
                Delete Task
              </Button>
            </ButtonGroup>
          }
        >
          {!!error && error}
          {body}
        </Card>
      </Modal>
    )
  }
}

ConfirmModal.displayName = "ConfirmModal"

ConfirmModal.propTypes = {
  body: PropTypes.node,
  getRequestArgs: PropTypes.func.isRequired,
  onFailure: PropTypes.func,
  onSuccess: PropTypes.func,
  renderPopup: PropTypes.func.isRequired,
  requestFn: PropTypes.func.isRequired,
  title: PropTypes.node,
  unrenderModal: PropTypes.func.isRequired,
  unrenderPopup: PropTypes.func.isRequired,
}

ConfirmModal.defaultProps = {
  body: "Are you sure?",
  getRequestArgs: () => null,
  onFailure: () => {},
  onSuccess: () => {},
  title: "Confirm",
}

export default props => (
  <ModalContext.Consumer>
    {mCtx => (
      <PopupContext.Consumer>
        {pCtx => (
          <ConfirmModal
            {...props}
            renderPopup={pCtx.renderPopup}
            unrenderPopup={pCtx.unrenderPopup}
            unrenderModal={mCtx.unrenderModal}
          />
        )}
      </PopupContext.Consumer>
    )}
  </ModalContext.Consumer>
)
