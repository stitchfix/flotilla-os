import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { get } from "lodash"
import axios from "axios"
import Button from "./Button"
import Card from "./Card"
import Modal from "./Modal"
import Popup from "./Popup"
import modalActions from "../actions/modalActions"
import popupActions from "../actions/popupActions"
import intentTypes from "../constants/intentTypes"
import config from "../config"

export class StopRunModal extends Component {
  static propTypes = {
    definitionId: PropTypes.string,
    dispatch: PropTypes.func,
    runId: PropTypes.string,
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
