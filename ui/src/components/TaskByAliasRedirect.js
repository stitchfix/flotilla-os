import React, { Component } from "react"
import { Redirect } from "react-router-dom"
import { connect } from "react-redux"
import { withStateFetch } from "aa-ui-components"
import { View } from "aa-ui-components"
import config from "../config"

export class TaskByAliasRedirect extends Component {
  constructor(props) {
    super(props)
    this.fetch = this.fetch.bind(this)
  }
  componentDidMount() {
    const id = this.props.match.params.alias
    this.fetch(id)
  }
  componentWillReceiveProps(nextProps) {
    if (this.props.match.params.alias !== nextProps.match.params.alias) {
      this.fetch(nextProps.match.params.definitionId)
    }
  }
  fetch(alias) {
    this.props.fetch(`${config.FLOTILLA_API}/task/alias/${alias}`)
  }
  render() {
    const { isLoading, data, error, match, dispatch } = this.props

    if (data && data.definition_id) {
      return <Redirect to={`/tasks/${data.definition_id}`} />
    }
    if (error) {
      if (error.response && error.response.status == 404) {
        return <View>No such task with alias "{match.params.alias}"</View>
      }
      return (
        <View>
          <pre>
            Error loading task with alias "{match.params.alias}":
            <br />
            {JSON.stringify(error, null, 2)}
          </pre>
        </View>
      )
    }
    if (isLoading) {
      return <View> Loading task with alias {match.params.alias} </View>
    }
    return <View> Waiting for data </View>
  }
}

export default connect()(withStateFetch(TaskByAliasRedirect))
