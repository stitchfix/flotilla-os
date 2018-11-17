import React, { Component } from "react"
import { withRouter } from "react-router-dom"
import { Form as ReactForm } from "react-form"
import { get, omit } from "lodash"
import Loader from "../styled/Loader"
import View from "../styled/View"
import Navigation from "../Navigation/Navigation"
import Form from "../Form/Form"
import FieldSelect from "../Form/FieldSelect"
import FieldKeyValue from "../Form/FieldKeyValue"
import api from "../../api"
import config from "../../config"
import * as requestStateTypes from "../../constants/requestStateTypes"
import TaskContext from "../Task/TaskContext"
import filterInvalidRunEnv from "../../utils/filterInvalidRunEnv"
import intentTypes from "../../constants/intentTypes"

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
    const { data, previousRunState } = this.props

    const cluster = get(
      previousRunState,
      "cluster",
      get(config, "DEFAULT_CLUSTER", "")
    )
    const env = filterInvalidRunEnv(
      get(previousRunState, "env", get(data, ["env"], []))
    )

    return {
      cluster,
      env,
      run_tags: get(config, "REQUIRED_RUN_TAGS", []).map(name => ({
        name,
        value: "",
      })),
    }
  }

  render() {
    const { requestState, definitionID, data, goBack } = this.props

    if (requestState === requestStateTypes.NOT_READY) {
      return <Loader />
    }

    const breadcrumbs = [
      { text: "Tasks", href: "/tasks" },
      {
        text: get(data, "alias", definitionID),
        href: `/tasks/${definitionID}`,
      },
      { text: "Run", href: `/tasks/${definitionID}/run` },
    ]

    const actions = [
      {
        isLink: false,
        text: "Cancel",
        buttonProps: {
          onClick: goBack,
        },
      },
      {
        isLink: false,
        text: "Run",
        buttonProps: {
          type: "submit",
          intent: intentTypes.primary,
        },
      },
    ]

    return (
      <ReactForm
        defaultValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
      >
        {formAPI => {
          return (
            <form onSubmit={formAPI.submitForm}>
              <View>
                <Navigation breadcrumbs={breadcrumbs} actions={actions} />
                <Form title={`Run ${get(data, "alias", definitionID)}`}>
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

export default withRouter(props => {
  return (
    <TaskContext.Consumer>
      {ctx => (
        <RunForm
          push={props.history.push}
          previousRunState={get(props, ["location", "state"], {})}
          goBack={props.history.goBack}
          {...omit(props, ["history", "location", "match", "staticContext"])}
          {...ctx}
        />
      )}
    </TaskContext.Consumer>
  )
})
