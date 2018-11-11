import React, { Component } from "react"
import PropTypes from "prop-types"
import { withRouter } from "react-router-dom"
import { connect } from "react-redux"
import { Form as ReactForm } from "react-form"
import { get, isEmpty, omit } from "lodash"

import Button from "../styled/Button"
import ButtonGroup from "../styled/ButtonGroup"
import Loader from "../styled/Loader"
import Popup from "../Popup/Popup"
import PopupContext from "../Popup/PopupContext"
import View from "../styled/View"
import ViewHeader from "../styled/ViewHeader"

import Form from "../Form/Form"
import FieldText from "../Form/FieldText"
import FieldSelect from "../Form/FieldSelect"
import FieldKeyValue from "../Form/FieldKeyValue"
import TaskContext from "../Task/TaskContext"
import api from "../../api"
import config from "../../config"

import * as requestStateTypes from "../../constants/requestStateTypes"

const taskFormTypes = {
  CREATE: "CREATE",
  UPDATE: "UPDATE",
  CLONE: "CLONE",
}

class TaskForm extends Component {
  static transformValues = values =>
    Object.keys(values).reduce((acc, k) => {
      if (k === "memory") {
        acc[k] = +values[k]
      } else {
        acc[k] = values[k]
      }

      return acc
    }, {})
  handleSubmit = values => {
    const { data, type, push, renderPopup } = this.props

    switch (type) {
      case taskFormTypes.UPDATE:
        api
          .updateTask({
            definitionID: get(data, "definition_id", ""),
            values: TaskForm.transformValues(values),
          })
          .then(responseData => {
            push(`/tasks/${get(responseData, "definition_id", "")}`)
          })
          .catch(error => {
            console.error(error)
          })
        break
      case taskFormTypes.CREATE:
      case taskFormTypes.CLONE:
        api
          .createTask({ values })
          .then(responseData => {
            push(`/tasks/${get(responseData, "definition_id", "")}`)
          })
          .catch(error => {
            console.error(error)
          })
        break
      default:
        console.warn("TaskForm's `type` prop was not specified, doing nothing.")
    }
  }

  renderTitle() {
    const { data, type } = this.props

    switch (type) {
      case taskFormTypes.CREATE:
        return "Create New Task"
      case taskFormTypes.UPDATE:
        return `Update ${get(data, "definition_id", "Task")}`
      case taskFormTypes.CLONE:
        return `Clone ${get(data, "definition_id", "Task")}`
      default:
        return "Task Form"
    }
  }

  shouldNotRenderForm() {
    const { type, groupOptions, tagOptions, requestState } = this.props

    if (isEmpty(groupOptions) || isEmpty(tagOptions)) {
      return true
    }

    if (
      type !== taskFormTypes.CREATE &&
      requestState === requestStateTypes.NOT_READY
    ) {
      return true
    }

    return false
  }

  getDefaultValues() {
    const { data, type } = this.props

    let ret = {
      memory: get(data, "memory", 1024),
      image: `${get(config, "IMAGE_PREFIX", "")}IMAGE_NAME:IMAGE_TAG`,
    }

    if (type === taskFormTypes.CREATE) {
      return ret
    }

    if (!isEmpty(data)) {
      return {
        ...ret,
        group_name: get(data, "group_name", ""),
        image: get(data, "image", ""),
        command: get(data, "command", ""),
        tags: get(data, "tags", []),
        env: get(data, "env", []),
      }
    }
  }

  render() {
    const { type, groupOptions, tagOptions, goBack } = this.props

    if (this.shouldNotRenderForm()) {
      return <Loader />
    }

    return (
      <ReactForm
        defaultValues={this.getDefaultValues()}
        onSubmit={this.handleSubmit}
      >
        {formAPI => {
          console.log(formAPI)
          return (
            <form onSubmit={formAPI.submitForm}>
              <View>
                <ViewHeader
                  title={this.renderTitle()}
                  actions={
                    <ButtonGroup>
                      <Button type="cancel" onClick={goBack}>
                        cancel
                      </Button>
                      <Button type="submit" intent="primary">
                        submit
                      </Button>
                    </ButtonGroup>
                  }
                />
                <Form>
                  {type !== taskFormTypes.UPDATE && (
                    <FieldText
                      label="Alias"
                      field="alias"
                      description="Choose a descriptive alias for this task."
                    />
                  )}
                  <FieldSelect
                    label="Group Name"
                    field="group_name"
                    options={groupOptions}
                    isCreatable
                    description="Create a new group name or select an existing one to help searching for this task in the future."
                  />
                  <FieldText
                    label="Image"
                    field="image"
                    description="The full URL of the Docker image and tag."
                  />
                  <FieldText
                    isTextArea
                    label="Command"
                    field="command"
                    description="The command for this task to execute."
                  />
                  <FieldText
                    isNumber
                    label="Memory (MB)"
                    field="memory"
                    description="The amount of memory this task needs."
                  />
                  <FieldSelect
                    isCreatable
                    isMulti
                    label="Tags"
                    field="tags"
                    options={tagOptions}
                  />
                  <FieldKeyValue
                    label="Environment Variables"
                    field="env"
                    addValue={formAPI.addValue}
                    removeValue={formAPI.removeValue}
                    values={get(formAPI, ["values", "env"], [])}
                    descripion="Environment variables that can be adjusted during runtime."
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

TaskForm.propTypes = {
  data: PropTypes.object,
  goBack: PropTypes.func.isRequired,
  groupOptions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    })
  ),
  push: PropTypes.func.isRequired,
  renderPopup: PropTypes.func.isRequired,
  requestState: PropTypes.oneOf(Object.values(requestStateTypes)),
  tagOptions: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    })
  ),
  type: PropTypes.oneOf(Object.values(taskFormTypes)).isRequired,
}

const mapStateToProps = state => ({
  groupOptions: get(state, ["selectOpts", "group"], []),
  tagOptions: get(state, ["selectOpts", "tag"], []),
})

const ReduxConnectedTaskForm = connect(mapStateToProps)(TaskForm)
const ConnectedTaskForm = withRouter(props => (
  <PopupContext.Consumer>
    {ctx => (
      <ReduxConnectedTaskForm
        {...omit(props, ["history", "location", "match", "staticContext"])}
        push={props.history.push}
        goBack={props.history.goBack}
        renderPopup={ctx.renderPopup}
      />
    )}
  </PopupContext.Consumer>
))

export const CreateTaskForm = () => (
  <ConnectedTaskForm type={taskFormTypes.CREATE} />
)

export const UpdateTaskForm = props => (
  <TaskContext.Consumer>
    {ctx => <ConnectedTaskForm type={taskFormTypes.UPDATE} {...ctx} />}
  </TaskContext.Consumer>
)

export const CloneTaskForm = props => (
  <TaskContext.Consumer>
    {ctx => <ConnectedTaskForm type={taskFormTypes.CLONE} {...ctx} />}
  </TaskContext.Consumer>
)
