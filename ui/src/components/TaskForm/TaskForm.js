import React, { Component, Fragment } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Form as ReactForm } from "react-form"
import { get, isEmpty } from "lodash"

import Button from "../Button"
import Loader from "../Loader"
import View from "../View"
import ViewHeader from "../ViewHeader"

import Form from "../Form/Form"
import FieldText from "../Form/FieldText"
import FieldSelect from "../Form/FieldSelect"
import FieldKeyValue from "../Form/FieldKeyValue"
import TaskContext from "../Task/TaskContext"
import api from "../../api"

import * as requestStateTypes from "../../constants/requestStateTypes"

const taskFormTypes = {
  CREATE: "CREATE",
  UPDATE: "UPDATE",
  CLONE: "CLONE",
}

class TaskForm extends Component {
  handleSubmit = values => {
    const { taskDefinition, type } = this.props

    switch (type) {
      case taskFormTypes.UPDATE:
        api
          .updateTask({
            definitionID: get(taskDefinition, "definition_id", ""),
            values,
          })
          .then(res => ({
            // Go back to task definition
          }))
          .catch(err => {
            // handle err
          })
        break
      case taskFormTypes.CREATE:
      case taskFormTypes.CLONE:
        api
          .createTask({ values })
          .then(res => ({
            // Go to task definition
          }))
          .catch(err => {
            // handle err
          })
        break
      default:
        console.warn("TaskForm's `type` prop was not specified, doing nothing.")
    }
  }

  renderTitle() {
    const { taskDefinition, type } = this.props

    switch (type) {
      case taskFormTypes.CREATE:
        return "Create New Task"
      case taskFormTypes.UPDATE:
        return `Update ${get(taskDefinition, "definition_id", "Task")}`
      case taskFormTypes.CLONE:
        return `Clone ${get(taskDefinition, "definition_id", "Task")}`
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
        env: get(data, "env", []).map(e => ({ key: e.name, value: e.value })),
      }
    }
  }

  render() {
    const { type, groupOptions, tagOptions } = this.props

    if (this.shouldNotRenderForm()) {
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
                  title={this.renderTitle()}
                  actions={
                    <Button type="submit" intent="primary">
                      submit
                    </Button>
                  }
                />
                <Form>
                  {type !== taskFormTypes.UPDATE && (
                    <FieldText label="Alias" field="alias" />
                  )}
                  <FieldSelect
                    label="Group Name"
                    field="group_name"
                    options={groupOptions}
                    isCreatable
                  />
                  <FieldText label="Image" field="image" />
                  <FieldText isTextArea label="Command" field="command" />
                  <FieldText isNumber label="Memory" field="memory" />
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
}

const mapStateToProps = state => ({
  groupOptions: get(state, ["selectOpts", "group"], []),
  tagOptions: get(state, ["selectOpts", "tag"], []),
})

const ConnectedTaskForm = connect(mapStateToProps)(TaskForm)

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
