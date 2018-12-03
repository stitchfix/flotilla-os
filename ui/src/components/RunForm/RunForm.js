import React, { Component } from "react"
import { withRouter } from "react-router-dom"
import { Form as ReactForm } from "react-form"
import { get, omit, isEmpty, has, intersection } from "lodash"
import Loader from "../styled/Loader"
import View from "../styled/View"
import Navigation from "../Navigation/Navigation"
import Form from "../styled/Form"
import { ReactFormFieldSelect } from "../Field/FieldSelect"
import ReactFormKVField from "../Field/ReactFormKVField"
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

  shouldDisableSubmitButton = formAPI => {
    if (!isEmpty(formAPI.errors)) {
      return true
    }

    const requiredValues = ["cluster"]

    for (let i = 0; i < requiredValues.length; i++) {
      if (!has(formAPI.values, requiredValues[i])) {
        return true
      }
    }

    const requiredRunTags = get(config, "REQUIRED_RUN_TAGS", [])
    const runTagsValues = get(formAPI, ["values", "run_tags"], [])

    if (requiredRunTags.length > 0) {
      if (
        intersection(runTagsValues.map(r => r.name), requiredRunTags).length ===
        0
      ) {
        return true
      }

      for (let i = 0; i < runTagsValues.length; i++) {
        if (
          requiredRunTags.includes(runTagsValues[i].name) &&
          !runTagsValues[i].value
        ) {
          return true
        }
      }
    }

    return false
  }

  getRunTagsDescription = () => {
    const requiredRunTags = get(config, "REQUIRED_RUN_TAGS", [])

    if (requiredRunTags.length < 1) {
      return null
    }

    return `The following run tags must be filled out in order for this task to run: ${requiredRunTags}.`
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

    return (
      <ReactForm
        defaultValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
      >
        {formAPI => {
          const shouldDisableSubmitButton = this.shouldDisableSubmitButton(
            formAPI
          )
          const actions = [
            {
              isLink: false,
              text: "Cancel",
              buttonProps: {
                onClick: goBack,
                type: "button",
              },
            },
            {
              isLink: false,
              text: "Run",
              buttonProps: {
                type: "submit",
                intent: intentTypes.primary,
                isDisabled: shouldDisableSubmitButton,
              },
            },
          ]

          return (
            <form onSubmit={formAPI.submitForm}>
              <View>
                <Navigation breadcrumbs={breadcrumbs} actions={actions} />
                <Form title={`Run ${get(data, "alias", definitionID)}`}>
                  <ReactFormFieldSelect
                    label="Cluster"
                    field="cluster"
                    requestOptionsFn={api.getClusters}
                    shouldRequestOptions
                    description="Select a cluster for this task to be executed on."
                    isRequired
                    validate={value =>
                      !value ? { error: "Value must not be null." } : null
                    }
                  />
                  <ReactFormKVField
                    label="Run Tags"
                    field="run_tags"
                    addValue={formAPI.addValue}
                    removeValue={formAPI.removeValue}
                    values={get(formAPI, ["values", "run_tags"], [])}
                    description={this.getRunTagsDescription()}
                  />
                  <ReactFormKVField
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
