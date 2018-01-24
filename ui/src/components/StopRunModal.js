import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import {
  Button,
  Card,
  Modal,
  modalActions,
  intentTypes,
  popupActions,
  Popup,
} from "aa-ui-components"
import { get } from "lodash"
import axios from "axios"
import config from "../config"

export class StopRunModal extends Component {
  static propTypes = {
    definitionId: PropTypes.string,
    runId: PropTypes.string,
    dispatch: PropTypes.func,
  }
  constructor(props) {
    super(props)
    this.handleStopButtonClick = this.handleStopButtonClick.bind(this)
  }
  state = {
    inFlight: false,
    error: false,
  }
  handleStopButtonClick() {
    const { definitionId, runId } = this.props

    this.setState({ inFlight: true })

    return axios
      .delete(`${config.FLOTILLA_API}/task/${definitionId}/history/${runId}`)
      .then(res => {
        this.setState({ inFlight: false })
        this.props.dispatch(modalActions.unrenderModal())
        this.props.dispatch(
          popupActions.renderPopup(
            <Popup
              title="Success!"
              message="Your run was stopped."
              intent={intentTypes.success}
              hide={() => {
                this.props.dispatch(popupActions.unrenderPopup())
              }}
            />
          )
        )
      })
      .catch(err => {
        const errorMessage = get(err, "response.data.error", err.toString())
        this.setState({
          inFlight: false,
          error: errorMessage,
        })
        this.props.dispatch(
          popupActions.renderPopup(
            <Popup
              title="Error!"
              message={errorMessage}
              intent={intentTypes.error}
              autohide={false}
              hide={() => {
                this.props.dispatch(popupActions.unrenderPopup())
              }}
            />
          )
        )
      })
  }
  render() {
    const { error, inFlight } = this.state
    const { dispatch } = this.props

    return (
      <Modal>
        <Card
          header="Confirm Stop Run"
          footer={
            <div className="flex with-horizontal-child-margin">
              <Button
                onClick={() => {
                  dispatch(modalActions.unrenderModal())
                }}
              >
                Cancel
              </Button>
              <Button
                intent={intentTypes.error}
                onClick={this.handleStopButtonClick}
                isLoading={inFlight}
              >
                Stop Run
              </Button>
            </div>
          }
        >
          Are you sure you want to stop this run?
        </Card>
      </Modal>
    )
  }
}

export default connect()(StopRunModal)
