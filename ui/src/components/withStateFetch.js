import React, { Component } from "react"
import axios from "axios"

const withStateFetch = UnwrappedComponent =>
  class WrappedComponent extends Component {
    static displayName = `withStateFetch(${UnwrappedComponent.displayName ||
      "UnwrappedComponent"})`
    constructor(props) {
      super(props)
      this.fetch = this.fetch.bind(this)
    }
    state = {
      isLoading: false,
      error: false,
      data: undefined,
    }
    fetch(url) {
      this.setState({ isLoading: true })
      return axios
        .get(url)
        .then(({ data }) => {
          this.setState({ isLoading: false, data })
        })
        .catch(error => {
          console.log(error)
          this.setState({ isLoading: false, error })
        })
    }
    render() {
      return (
        <UnwrappedComponent
          {...this.state}
          {...this.props}
          fetch={this.fetch}
        />
      )
    }
  }

export default withStateFetch
