import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import { connect } from "react-redux"
import { Creatable } from "react-select"
import { reduxForm } from "redux-form"
import Helmet from "react-helmet"
import { get, isEmpty, has } from "lodash"
import {
  ReduxFormGroupSelect,
  View,
  ViewHeader,
  Card,
  withRouterSync,
  Button,
  intentTypes,
} from "aa-ui-components"
import config from "../config"
import { envNameValueDelimiterChar } from "../constants"
import { runFormValidate, getHelmetTitle } from "../utils/"
import withFormSubmitter from "./withFormSubmitter"
import EnvFieldArray from "./EnvFieldArray"

export class RunForm extends Component {
  static displayName = "RunForm"
  static propTypes = {
    data: PropTypes.object,
    inFlight: PropTypes.bool,
    error: PropTypes.any,
    handleSubmit: PropTypes.func,
  }
  constructor(props) {
    super(props)
    this.handleEnvCreate = this.handleEnvCreate.bind(this)
    this.handleEnvUpdate = this.handleEnvUpdate.bind(this)
    this.handleEnvRemove = this.handleEnvRemove.bind(this)
  }
  state = {
    didSetQuery: false,
  }
  componentDidMount() {
    const { didSetQuery } = this.state
    const { data, query } = this.props

    if (this.shouldSetQuery()) {
      this.setQuery()
    }
  }
  componentDidUpdate(prevProps) {
    const { didSetQuery } = this.state
    const { data, query } = this.props

    if (this.shouldSetQuery()) {
      this.setQuery()
    }
  }
  // Ensure the following conditions are met before setting the query, and
  // thus, the initial values: 1. there should not be an existing query in
  // the URL, 2. this.state.didSetQuery should be false, 3. data should have
  // been fetched.
  shouldSetQuery() {
    const { didSetQuery } = this.state
    const { data, query } = this.props

    return isEmpty(query) && !didSetQuery && !isEmpty(data)
  }
  setQuery() {
    const { didSetQuery } = this.state
    const { data, updateQuery, query } = this.props

    if (isEmpty(query) && !didSetQuery && !isEmpty(data)) {
      this.setState({ didSetQuery: true }, () => {
        let updates = [
          {
            key: "cluster",
            value: config.DEFAULT_CLUSTER,
            updateType: "SHALLOW",
            replace: true,
          },
        ]

        if (Array.isArray(data.env) && data.env.length > 0) {
          updates = [
            ...updates,
            ...data.env.map(e => ({
              key: "env",
              value: `${e.name}|${e.value}`,
              updateType: "DEEP_CREATE",
              replace: true,
            })),
          ]
        }

        updateQuery(updates)
      })
    }
  }
  handleEnvCreate() {
    this.props.updateQuery({
      key: "env",
      value: envNameValueDelimiterChar,
      updateType: "DEEP_CREATE",
      replace: true,
    })
  }
  handleEnvRemove(index) {
    this.props.updateQuery({
      key: "env",
      updateType: "DEEP_REMOVE",
      index,
      replace: true,
    })
  }
  handleEnvUpdate({ nameOrValue, value, index }) {
    let split

    // Determine if there are multiple env values
    if (Array.isArray(this.props.query.env)) {
      split = this.props.query.env[index].split(envNameValueDelimiterChar)
    } else {
      split = this.props.query.env.split(envNameValueDelimiterChar)
    }

    let nextVal

    if (nameOrValue === "name") {
      nextVal = `${value}|${split[1]}`
    } else if (nameOrValue === "value") {
      nextVal = `${split[0]}|${value}`
    }

    this.props.updateQuery({
      key: "env",
      value: nextVal,
      updateType: "DEEP_UPDATE",
      index,
      replace: true,
    })
  }
  getTitle() {
    const { data, definitionId } = this.props

    return (
      <span>
        Run{" "}
        <Link to={`/tasks/${definitionId}`}>
          {get(data, "alias", definitionId)}
        </Link>
      </span>
    )
  }
  render() {
    const {
      inFlight,
      data,
      error,
      handleSubmit,
      definitionId,
      invalid,
      history,
    } = this.props

    return (
      <form onSubmit={handleSubmit}>
        <View>
          <Helmet>
            <title>
              {getHelmetTitle(`Run ${get(data, "alias", definitionId)}`)}
            </title>
          </Helmet>
          <ViewHeader
            title={this.getTitle()}
            actions={
              <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
                <Button
                  onClick={() => {
                    history.goBack()
                  }}
                  type="button"
                >
                  Cancel
                </Button>
                <Button
                  isLoading={inFlight}
                  intent={intentTypes.primary}
                  type="submit"
                  disabled={invalid}
                >
                  Run
                </Button>
              </div>
            }
          />
          <div className="flex ff-rn j-c a-c full-width">
            <Card
              containerStyle={{ maxWidth: 600 }}
              contentStyle={{ padding: 0 }}
            >
              <div className="key-value-container vertical full-width">
                <ReduxFormGroupSelect
                  name="cluster"
                  label="Cluster"
                  isRequired
                  options={this.props.clusterOptions}
                  onChange={(evt, value) => {
                    this.props.updateQuery({
                      key: "cluster",
                      value,
                      updateType: "SHALLOW",
                    })
                  }}
                />
                <EnvFieldArray
                  handleEnvCreate={this.handleEnvCreate}
                  handleEnvRemove={this.handleEnvRemove}
                  handleEnvUpdate={this.handleEnvUpdate}
                />
              </div>
            </Card>
          </div>
        </View>
      </form>
    )
  }
}

// A helper function to convert a string
const stringToEnvObject = str => {
  const split = str.split(envNameValueDelimiterChar)
  return {
    name: split[0],
    value: split[1],
  }
}

const mapStateToProps = (state, ownProps) => {
  const ret = {
    clusterOptions: get(state, "selectOpts.cluster", []),
  }

  // Populate form's initial values with props.query.
  if (!isEmpty(ownProps.query)) {
    const initialValues = {}
    if (has(ownProps.query, "cluster")) {
      initialValues.cluster = ownProps.query.cluster
    }

    if (has(ownProps.query, "env")) {
      if (Array.isArray(ownProps.query.env)) {
        initialValues.env = ownProps.query.env.map(stringToEnvObject)
      } else {
        initialValues.env = [stringToEnvObject(ownProps.query.env)]
      }
    }

    if (!isEmpty(initialValues)) {
      ret.initialValues = initialValues
    }
  }

  return ret
}

// Note: withRouterSync should be the outer-most wrapper, otherwise
// query updates will not trigger a rerender, most likely due to
// `connect` or `reduxForm` returning `false` in their
// shouldComponentUpdate lifecycle methods. See:
// https://github.com/ReactTraining/react-router/blob/master/packages/react-router/docs/guides/blocked-updates.md
export default withRouterSync(
  withFormSubmitter({
    getUrl: props =>
      `${config.FLOTILLA_API}/task/${props.definitionId}/execute`,
    httpMethod: "put",
    headers: { "content-type": "application/json" },
    onSuccess: (props, res) => {
      props.history.push(`/runs/${res.run_id}`)
    },
    onFailure: (props, err) => {
      console.error(err)
    },
  })(
    connect(mapStateToProps)(
      reduxForm({
        form: "run",
        validate: runFormValidate,
      })(RunForm)
    )
  )
)
