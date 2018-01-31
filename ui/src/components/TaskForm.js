import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Link } from "react-router-dom"
import { reduxForm } from "redux-form"
import Helmet from "react-helmet"
import { get } from "lodash"
import {
  ReduxFormGroupInput,
  ReduxFormGroupSelect,
  ReduxFormGroupTextarea,
  Card,
  View,
  ViewHeader,
  Button,
  intentTypes,
} from "aa-ui-components"
import config from "../config"
import { taskFormTypes } from "../constants/"
import { taskFormUtils, getHelmetTitle } from "../utils/"
import EnvFieldArray from "./EnvFieldArray"

export class TaskForm extends Component {
  static propTypes = {
    taskFormType: PropTypes.oneOf(Object.values(taskFormTypes)),
    handleSubmit: PropTypes.func,
    inFlight: PropTypes.bool,
    groupOptions: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.string,
      })
    ),
    tagOptions: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.string,
      })
    ),
    selectOptionsRequestInFlight: PropTypes.bool,
  }
  getTitle() {
    const { taskFormType, definitionId, data } = this.props

    switch (taskFormType) {
      case taskFormTypes.create:
        return "Create New Task"
      case taskFormTypes.edit:
        return `Edit ${get(data, "alias", definitionId)}`
      case taskFormTypes.copy:
        return `Copy ${get(data, "alias", definitionId)}`
      default:
        return ""
    }
  }
  render() {
    const {
      handleSubmit,
      inFlight,
      groupOptions,
      tagOptions,
      selectOptionsRequestInFlight,
      invalid,
      history,
    } = this.props

    return (
      <form onSubmit={handleSubmit} className="flex ff-rn j-c a-c full-width">
        <View>
          <Helmet>
            <title>{getHelmetTitle(this.getTitle())}</title>
          </Helmet>
          <ViewHeader
            title={this.getTitle()}
            actions={
              <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
                <Button
                  type="button"
                  onClick={() => {
                    history.goBack()
                  }}
                >
                  Cancel
                </Button>
                <Button
                  isLoading={inFlight}
                  intent={intentTypes.primary}
                  type="submit"
                  disabled={invalid}
                >
                  Save
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
                {this.props.taskFormType !== taskFormTypes.edit && (
                  <ReduxFormGroupInput name="alias" label="Alias" isRequired />
                )}
                <ReduxFormGroupSelect
                  name="group_name"
                  label="Group Name"
                  isRequired
                  options={groupOptions}
                  allowCreate
                  isLoading={selectOptionsRequestInFlight}
                  disabled={selectOptionsRequestInFlight}
                />
                <ReduxFormGroupInput name="image" label="Image" isRequired />
                <ReduxFormGroupTextarea
                  name="command"
                  label="Command"
                  isRequired
                />
                <ReduxFormGroupInput
                  name="memory"
                  label="Memory"
                  isRequired
                  type="number"
                />
                <ReduxFormGroupSelect
                  name="tags"
                  label="Tags"
                  options={tagOptions}
                  multi
                  allowCreate
                  isLoading={selectOptionsRequestInFlight}
                  disabled={selectOptionsRequestInFlight}
                />
                <EnvFieldArray />
              </div>
            </Card>
          </div>
        </View>
      </form>
    )
  }
}

export default connect(taskFormUtils.mapStateToProps)(
  reduxForm({
    form: "task",
    validate: taskFormUtils.validate,
  })(TaskForm)
)
