import React, { Component } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { reduxForm } from "redux-form"
import Helmet from "react-helmet"
import { get } from "lodash"
import Button from "./Button"
import Card from "./Card"
import EnvFieldArray from "./EnvFieldArray"
import ReduxFormGroupInput from "./ReduxFormGroupInput"
import ReduxFormGroupSelect from "./ReduxFormGroupSelect"
import ReduxFormGroupTextarea from "./ReduxFormGroupTextarea"
import View from "./View"
import ViewHeader from "./ViewHeader"
import intentTypes from "../constants/intentTypes"
import taskFormTypes from "../constants/taskFormTypes"
import getHelmetTitle from "../utils/getHelmetTitle"
import taskFormUtils from "../utils/taskFormUtils"

export class TaskForm extends Component {
  static propTypes = {
    groupOptions: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.string,
      })
    ),
    handleSubmit: PropTypes.func,
    inFlight: PropTypes.bool,
    selectOptionsRequestInFlight: PropTypes.bool,
    tagOptions: PropTypes.arrayOf(
      PropTypes.shape({
        label: PropTypes.string,
        value: PropTypes.string,
      })
    ),
    taskFormType: PropTypes.oneOf(Object.values(taskFormTypes)),
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
