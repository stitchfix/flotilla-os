import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import modalActions from "../actions/modalActions"

const modalHOC = UnwrappedComponent => {
  class WrappedComponent extends Component {
    static displayName = `modalHOC(${UnwrappedComponent.displayName ||
      "UnwrappedComponent"})`
    static propTypes = {
      dispatch: PropTypes.func,
    }
    constructor(props) {
      super(props)
      this.unrenderModal = this.unrenderModal.bind(this)
      this.renderModal = this.renderModal.bind(this)
    }
    unrenderModal() {
      this.props.dispatch(modalActions.unrenderModal())
    }
    renderModal(modal) {
      this.props.dispatch(modalActions.renderModal(modal))
    }
    render() {
      return (
        <UnwrappedComponent
          renderModal={this.renderModal}
          unrenderModal={this.unrenderModal}
          {...this.props}
        />
      )
    }
  }

  return connect()(WrappedComponent)
}

export default modalHOC
