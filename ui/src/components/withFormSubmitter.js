import React, { Component } from "react"
import { connect } from "react-redux"
import axios from "axios"
import { isFunction, get } from "lodash"
import intentTypes from "../constants/intentTypes"
import popupActions from "../actions/popupActions"
import Popup from "./Popup"

const validateOptions = options => {
  if (!options.getUrl) {
    console.error("You must pass a `getUrl` option to withFormSubmitter.")
    return false
  }

  if (typeof options.getUrl !== "function") {
    console.error(
      "The `getUrl` option passed to withFormSubmitter must be a function."
    )
    return false
  }

  if (!options.httpMethod) {
    console.error("You must pass a `httpMethod` option to withFormSubmitter.")
    return false
  }

  if (
    options.httpMethod.toUpperCase() !== "POST" &&
    options.httpMethod.toUpperCase() !== "PUT"
  ) {
    console.error(
      "The `httpMethod` option passed to withFormSubmitter must be a 'POST' or 'PUT'."
    )
    return false
  }

  if (
    !!options.transformFormValues &&
    typeof options.transformFormValues !== "function"
  ) {
    console.error(
      "The `transformFormValues` option passed to withFormSubmitter must be a function."
    )
    return false
  }

  return true
}

export default function withFormSubmitter(options) {
  validateOptions(options)

  const {
    getUrl,
    httpMethod,
    onFailure = () => {},
    onSuccess = () => {},
    transformFormValues,
  } = options

  return UnwrappedComponent => {
    class WrappedComponent extends Component {
      constructor(props) {
        super(props)
        this.handleSubmit = this.handleSubmit.bind(this)
      }
      state = {
        inFlight: false,
        error: false,
      }
      handleSubmit(formValues) {
        this.setState({ inFlight: true })
        let _formValues = formValues

        // Run the form values through a transformation function, if necessary.
        if (isFunction(transformFormValues)) {
          _formValues = transformFormValues(formValues)
        }

        axios({
          url: getUrl(this.props),
          method: httpMethod,
          data: _formValues,
        })
          .then(({ data }) => {
            this.setState({ inFlight: false })
            onSuccess(this.props, data)

            this.props.dispatch(
              popupActions.renderPopup(
                <Popup
                  title="Submitted successfully!"
                  intent={intentTypes.success}
                  hide={() => {
                    this.props.dispatch(popupActions.unrenderPopup())
                  }}
                />
              )
            )
          })
          .catch(error => {
            console.error(error)
            this.setState({ inFlight: false, error })

            onFailure(this.props, error)
            this.props.dispatch(
              popupActions.renderPopup(
                <Popup
                  title="Submit error occurred."
                  message={get(error, "response.data.error", error.toString())}
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
        return (
          <UnwrappedComponent
            {...this.props}
            onSubmit={this.handleSubmit}
            inFlight={this.state.inFlight}
            error={this.state.error}
          />
        )
      }
    }

    return connect()(WrappedComponent)
  }
}
