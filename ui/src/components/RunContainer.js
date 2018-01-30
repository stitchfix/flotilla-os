import React, { Component } from "react"
import PropTypes from "prop-types"
import { HashRouter, Route, Switch } from "react-router-dom"
import { get } from "lodash"
import { withStateFetch } from "aa-ui-components"
import config from "../config"
import { runStatusTypes } from "../constants"
import RunView from "./RunView"
import RunMiniView from "./RunMiniView"

const interval = 5000
const getUrl = runId => `${config.FLOTILLA_API}/task/history/${runId}`

export class RunContainer extends Component {
  static propTypes = {
    match: PropTypes.shape({
      params: PropTypes.shape({
        runId: PropTypes.string.isRequired,
      }),
    }),
    data: PropTypes.object,
    isLoading: PropTypes.bool,
    error: PropTypes.any,
  }
  constructor(props) {
    super(props)
  }
  componentDidMount() {
    this.fetch(getUrl(this.props.match.params.runId))
    this.startInterval()
  }
  componentWillReceiveProps(nextProps) {
    if (get(nextProps.data, "status", false) === runStatusTypes.stopped) {
      this.stopInterval()
    }
    if (this.props.match.params.runId !== nextProps.match.params.runId) {
      this.stopInterval()
      this.fetch(getUrl(nextProps.match.params.runId))
      this.startInterval()
    }
  }
  componentWillUnmount() {
    this.stopInterval()
  }
  fetch(url) {
    this.props.fetch(url)
  }
  startInterval() {
    this.interval = window.setInterval(() => {
      this.fetch(getUrl(this.props.match.params.runId))
    }, interval)
  }
  stopInterval() {
    window.clearInterval(this.interval)
  }
  render() {
    const { isLoading, data, error, match } = this.props
    const rootPath = match.url

    return (
      <HashRouter>
        <Switch>
          <Route
            exact
            path={rootPath}
            render={() => (
              <RunView
                runId={match.params.runId}
                isLoading={isLoading}
                data={data}
                error={error}
              />
            )}
          />
          <Route
            exact
            path={`${rootPath}/mini`}
            render={() => (
              <RunMiniView
                runId={match.params.runId}
                isLoading={isLoading}
                data={data}
                error={error}
              />
            )}
          />
        </Switch>
      </HashRouter>
    )
  }
}

export default withStateFetch(RunContainer)
