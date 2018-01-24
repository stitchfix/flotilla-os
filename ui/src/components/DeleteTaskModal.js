import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { withRouter } from "react-router-dom"
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

export class DeleteTaskModal extends Component {
  static propTypes = {
    definitionId: PropTypes.string,
    dispatch: PropTypes.func,
    history: PropTypes.shape({
      push: PropTypes.func,
    }),
  }
  constructor(props) {
    super(props)
    this.handleDeleteButtonClick = this.handleDeleteButtonClick.bind(this)
  }
  state = {
    inFlight: false,
    error: false,
  }
  handleDeleteButtonClick() {
    const { definitionId } = this.props

    this.setState({ inFlight: true })

    return axios
      .delete(`${config.FLOTILLA_API}/task/${definitionId}`)
      .then(res => {
        this.setState({ inFlight: false })
        this.props.dispatch(
          popupActions.renderPopup(
            <Popup
              title="Success!"
              message="Your task was deleted."
              intent={intentTypes.success}
              hide={() => {
                this.props.dispatch(popupActions.unrenderPopup())
              }}
            />
          )
        )
        this.props.dispatch(modalActions.unrenderModal())
        this.props.history.push("/tasks")
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
          header="Confirm"
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
                onClick={this.handleDeleteButtonClick}
                isLoading={inFlight}
              >
                Delete Task
              </Button>
            </div>
          }
        >
          Are you sure you want to delete this task?
        </Card>
      </Modal>
    )
  }
}

export default withRouter(connect()(DeleteTaskModal))
