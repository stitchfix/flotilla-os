import React, { Component } from "react"
import PropTypes from "prop-types"
import { withRouter } from "react-router-dom"
import { Form as ReactForm } from "react-form"
import { get, isEmpty, omit } from "lodash"
import Button from "../styled/Button"
import Loader from "../styled/Loader"
import View from "../styled/View"
import ViewHeader from "../styled/ViewHeader"
import Form from "../Form/Form"
import FieldSelect from "../Form/FieldSelect"
import FieldKeyValue from "../Form/FieldKeyValue"
import api from "../../api"
import config from "../../config"

import * as requestStateTypes from "../../constants/requestStateTypes"

import TaskContext from "../Task/TaskContext"

class RunForm extends Component {
  static transformRunTags = arr =>
    arr.reduce((acc, val) => {
      acc[val.name] = val.value
      return acc
    }, {})

  handleSubmit = values => {
    const { data, push } = this.props

    api
      .runTask({
        values: {
          ...values,
          run_tags: RunForm.transformRunTags(values.run_tags),
        },
        definitionID: data.definition_id,
      })
      .then(res => {
        push(`/runs/${res.run_id}`)
      })
      .catch(error => {
        console.log(error)
      })
  }

  getDefaultValues = () => {
    const { data } = this.props

    return {
      cluster: get(config, "DEFAULT_CLUSTER", ""),
      run_tags: [{ name: "owner_id", value: "" }],
      env: get(data, ["env"], []),
    }
  }

  render() {
    const { requestState } = this.props

    if (requestState === requestStateTypes.NOT_READY) {
      return <Loader />
    }

    return (
      <ReactForm
        defaultValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
      >
        {formAPI => {
          return (
            <form onSubmit={formAPI.submitForm}>
              <View>
                <ViewHeader
                  title="fill me out"
                  actions={
                    <Button type="submit" intent="primary">
                      submit
                    </Button>
                  }
                />
                <Form>
                  <FieldSelect
                    label="Cluster"
                    field="cluster"
                    requestOptionsFn={api.getClusters}
                    shouldRequestOptions
                  />
                  <FieldKeyValue
                    label="Run Tags"
                    field="run_tags"
                    addValue={formAPI.addValue}
                    removeValue={formAPI.removeValue}
                    values={get(formAPI, ["values", "run_tags"], [])}
                  />
                  <FieldKeyValue
                    label="Environment Variables"
                    field="env"
                    addValue={formAPI.addValue}
                    removeValue={formAPI.removeValue}
                    values={get(formAPI, ["values", "env"], [])}
                  />
                </Form>
              </View>
            </form>
          )
        }}
      </ReactForm>
    )
  }
}

RunForm.propTypes = {}

export default withRouter(props => (
  <TaskContext.Consumer>
    {ctx => (
      <RunForm
        push={props.history.push}
        {...omit(props, ["history", "location", "match", "staticContext"])}
        {...ctx}
      />
    )}
  </TaskContext.Consumer>
))
