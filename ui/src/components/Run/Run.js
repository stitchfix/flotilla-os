import React, { Component } from "react"
import PropTypes from "prop-types"
import { Switch, Route } from "react-router-dom"
import { get, omit, isEqual } from "lodash"
import * as requestStateTypes from "../../constants/requestStateTypes"
import api from "../../api"
import config from "../../config"
import RunContext from "./RunContext"
import RunView from "./RunView"
import runStatusTypes from "../../constants/runStatusTypes"
import PopupContext from "../Popup/PopupContext"
import intentTypes from "../../constants/intentTypes"

class Run extends Component {
  state = {
    inFlight: false,
    error: false,
    data: {},
    requestState: requestStateTypes.NOT_READY,
  }

  componentDidMount() {
    this.requestData()

    this.requestInterval = window.setInterval(() => {
      this.requestData()
    }, config.RUN_REQUEST_INTERVAL_MS)
  }

  componentDidUpdate(prevProps, prevState) {
    if (!isEqual(prevProps.runID, this.props.runID)) {
      this.requestData()
    }

    if (
      get(prevState, ["data", "status"]) !== runStatusTypes.stopped &&
      get(this.state, ["data", "status"]) === runStatusTypes.stopped
    ) {
      this.clearInterval()
    }
  }

  componentWillUnmount() {
    this.clearInterval()
  }

  clearInterval = () => {
    window.clearInterval(this.requestInterval)
  }

  requestData = () => {
    // If the previous request is still in flight, return.
    if (this.state.inFlight === true) {
      return
    }

    this.setState({ inFlight: false, error: false })

    api
      .getRun({ runID: this.props.runID })
      .then(data => {
        this.setState({
          inFlight: false,
          data,
          error: false,
          requestState: requestStateTypes.READY,
        })
      })
      .catch(error => {
        this.clearInterval()
        const e = error.getError()

        this.props.renderPopup({
          body: e.data,
          intent: intentTypes.error,
          shouldAutohide: false,
          title: `Error (${e.status})`,
        })

        this.setState({
          inFlight: false,
          error,
          requestState: requestStateTypes.ERROR,
        })
      })
  }

  getCtx() {
    const { runID } = this.props
    return {
      ...this.state,
      runID,
    }
  }

  render() {
    const { rootPath } = this.props

    return (
      <RunContext.Provider value={this.getCtx()}>
        <Switch>
          <Route exact path={rootPath} component={RunView} />
        </Switch>
      </RunContext.Provider>
    )
  }
}

Run.propTypes = {
  renderPopup: PropTypes.func.isRequired,
  rootPath: PropTypes.string.isRequired,
  runID: PropTypes.string.isRequired,
}

export default props => (
  <PopupContext.Consumer>
    {ctx => (
      <Run
        {...omit(props, ["history", "location", "match", "staticContext"])}
        runID={get(props, ["match", "params", "runID"], "")}
        rootPath={get(props, ["match", "url"], "")}
        renderPopup={ctx.renderPopup}
      />
    )}
  </PopupContext.Consumer>
)
