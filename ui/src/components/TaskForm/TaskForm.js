import React, { Component, Fragment } from "react"
import PropTypes from "prop-types"
import { connect } from "react-redux"
import { Form as ReactForm } from "react-form"
import { get, isEmpty } from "lodash"

import Button from "../Button"
import View from "../View"
import ViewHeader from "../ViewHeader"

import Form from "../Form/Form"
import FieldText from "../Form/FieldText"
import FieldSelect from "../Form/FieldSelect"
import FieldKeyValue from "../Form/FieldKeyValue"
import api from "../../api"

const taskFormTypes = {
  CREATE: "CREATE",
  UPDATE: "UPDATE",
  CLONE: "CLONE",
}

class TaskForm extends Component {
  handleSubmit = values => {
    console.log(values)
  }

  renderTitle() {}

  render() {
    const { type, groupOptions, tagOptions } = this.props

    if (isEmpty(groupOptions) || isEmpty(tagOptions)) {
      return "loadding"
    }

    return (
      <ReactForm
        // DEFAULT VALUES
        onSubmit={this.handleSubmit}
      >
        {formAPI => {
          return (
            <form onSubmit={formAPI.submitForm}>
              <View>
                <ViewHeader
                  title="FILL THIS OUT"
                  actions={
                    <Button type="submit" intent="primary">
                      submit
                    </Button>
                  }
                />
                <Form>
                  <FieldText label="Alias" field="alias" />
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
        {/* {fuck => {
            console.log(fuck)
            return (
              <Fragment>
                <FieldText label="Alias" field="alias" id={type} />
                
                
                <Button type="submit">submit</Button>
              </Fragment>
            )
          }} */}
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
  type: PropTypes.oneOf(Object.values(taskFormTypes)),
}

const mapStateToProps = state => ({
  groupOptions: get(state, ["selectOpts", "group"], []),
  tagOptions: get(state, ["selectOpts", "tag"], []),
})

const ConnectedTaskForm = connect(mapStateToProps)(TaskForm)

export const CreateTaskForm = () => (
  <ConnectedTaskForm type={taskFormTypes.CREATE} />
)
export const UpdateTaskForm = () => (
  <ConnectedTaskForm type={taskFormTypes.UPDATE} />
)
export const CloneTaskForm = () => (
  <ConnectedTaskForm type={taskFormTypes.CLONE} />
)
