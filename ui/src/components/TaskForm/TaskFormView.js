import React, { Component, Fragment } from "react"
import PropTypes from "prop-types"
import { Form } from "informed"

import Button from "../Button"
import View from "../View"
import ViewHeader from "../ViewHeader"

import FieldText from "../Form/FieldText"
import FieldNumber from "../Form/FieldNumber"
import FieldTextarea from "../Form/FieldTextarea"
import FieldSelect from "../Form/FieldSelect"
import FieldCreatableSelect from "../Form/FieldCreatableSelect"

import api from "../../api"

const taskFormTypes = {
  CREATE: "CREATE",
  UPDATE: "UPDATE",
  CLONE: "CLONE",
}

class TaskFormView extends Component {
  handleSubmit = values => {
    console.log(values)
  }

  renderTitle() {}

  render() {
    const { type } = this.props
    return (
      <View>
        <ViewHeader title="FILL THIS OUT" />
        <Form id={type} onSubmit={this.handleSubmit}>
          {({ formState }) => {
            console.log(formState)
            return (
              <Fragment>
                <FieldText label="Alias" field="alias" id={type} />
                <FieldCreatableSelect
                  label="Group Name"
                  field="group_name"
                  id={type}
                />
                <FieldText label="Image" field="image" id={type} />
                <FieldTextarea label="Command" field="command" id={type} />
                <FieldNumber label="Memory" field="memory" id={type} />
                <FieldCreatableSelect
                  isMulti
                  label="Tags"
                  field="tags"
                  id={type}
                />

                <Button type="submit">submit</Button>
              </Fragment>
            )
          }}
        </Form>
      </View>
    )
  }
}

TaskFormView.propTypes = {
  type: PropTypes.oneOf(Object.values(taskFormTypes)),
}

export const CreateTaskForm = () => <TaskFormView type={taskFormTypes.CREATE} />
export const UpdateTaskForm = () => <TaskFormView type={taskFormTypes.UPDATE} />
export const CloneTaskForm = () => <TaskFormView type={taskFormTypes.CLONE} />
