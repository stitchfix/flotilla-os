import React, { Component } from "react"
import PropTypes from "prop-types"
import Button from "../styled/Button"
import Card from "../styled/Card"
import ModalContext from "./ModalContext"
import Modal from "./Modal"
import PopupContext from "../Popup/PopupContext"
import intentTypes from "../../constants/intentTypes"

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
      onSuccess,
      onFailure,
      getRequestArgs,
    } = this.props

    this.setState({ inFlight: true, error: false })

    requestFn(getRequestArgs())
      .then(res => {
        renderPopup({
          body: "Action was completed successfully.",
          title: "Success!",
          intent: intentTypes.success,
        })
        unrenderModal()
        onSuccess(res)
      })
      .catch(error => {
        this.setState({ inFlight: false, error })

        renderPopup({
          body: "TODO: put error text here",
          title: "Error!",
          intent: intentTypes.error,
          shouldAutohide: false,
        })

        onFailure()
      })
  }

  render() {
    const { unrenderModal, body, title } = this.props
    const { inFlight, error } = this.state

    return (
      <Modal>
        <Card
          title={title}
          footerActions={[
            <Button onClick={unrenderModal}>Cancel</Button>,
            <Button
              intent={intentTypes.error}
              onClick={this.handleConfirm}
              isLoading={inFlight}
            >
              Delete Task
            </Button>,
          ]}
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
