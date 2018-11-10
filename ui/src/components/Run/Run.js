import React, { Component } from "react"
import PropTypes from "prop-types"
import { Switch, Route } from "react-router-dom"
import { get, omit, isEqual } from "lodash"
import * as requestStateTypes from "../../constants/requestStateTypes"
import api from "../../api"
import config from "../../config"
import RunContext from "./RunContext"
import RunView from "./RunView"

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

  componentDidUpdate(prevProps) {
    if (!isEqual(prevProps.runID, this.props.runID)) {
      this.requestData()
    }
  }

  componentWillUnmount() {
    window.clearInterval(this.requestInterval)
  }

  requestData = () => {
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
  rootPath: PropTypes.string.isRequired,
  runID: PropTypes.string.isRequired,
}

export default props => (
  <Run
    {...omit(props, ["history", "location", "match", "staticContext"])}
    runID={get(props, ["match", "params", "runID"], "")}
    rootPath={get(props, ["match", "url"], "")}
  />
)
