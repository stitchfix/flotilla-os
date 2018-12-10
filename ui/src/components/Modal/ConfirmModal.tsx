import * as React from "react"
import Button from "../styled/Button"
import ButtonGroup from "../styled/ButtonGroup"
import Card from "../styled/Card"
import ModalContext from "./ModalContext"
import Modal from "./Modal"
import PopupContext from "../Popup/PopupContext"
import { IPopupProps, intents } from "../../.."

interface IConfirmModalProps {
  body?: React.ReactNode
  getRequestArgs?: () => any
  onFailure?: (error: any) => void
  onSuccess?: (response: any) => void
  requestFn: (opts: any) => Promise<any>
  title?: React.ReactNode
}

interface IUnwrappedConfirmModalProps extends IConfirmModalProps {
  renderPopup: (popupProps: IPopupProps) => void
  unrenderModal: () => void
  unrenderPopup: () => void
}

interface IConfirmModalState {
  inFlight: boolean
  error: any
}

export class UnwrappedConfirmModal extends React.Component<
  IUnwrappedConfirmModalProps,
  IConfirmModalState
> {
  static defaultProps: Partial<IUnwrappedConfirmModalProps> = {
    body: "Are you sure?",
    getRequestArgs: () => null,
    onFailure: () => {},
    onSuccess: () => {},
    title: "Confirm",
  }

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

    requestFn(!!getRequestArgs ? getRequestArgs() : {})
      .then(res => {
        renderPopup({
          body: "Action was completed successfully.",
          title: "Success!",
          intent: intents.SUCCESS,
        })
        unrenderModal()
        if (onSuccess) onSuccess(res)
      })
      .catch(error => {
        this.setState({ inFlight: false, error: error.data })

        renderPopup({
          body: "TODO: put error text here",
          title: "Error!",
          intent: intents.ERROR,
          shouldAutohide: false,
        })

        if (onFailure) onFailure(error)
      })
  }

  render() {
    const { unrenderModal, body, title } = this.props
    const { inFlight, error } = this.state

    return (
      <Modal>
        <Card
          title={title}
          footerActions={
            <ButtonGroup>
              <Button id="cancel" onClick={unrenderModal}>
                Cancel
              </Button>
              <Button
                id="confirm"
                intent={intents.ERROR}
                onClick={this.handleConfirm}
                isLoading={inFlight}
              >
                Confirm
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

const ConfirmModal = (props: IConfirmModalProps) => (
  <ModalContext.Consumer>
    {mCtx => (
      <PopupContext.Consumer>
        {pCtx => (
          <UnwrappedConfirmModal
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

export default ConfirmModal
