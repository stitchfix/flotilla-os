import React from "react"
import { connect } from "react-redux"
import { withRouter } from "react-router-dom"
import config from "../config"
import { taskFormTypes } from "../constants"
import { taskFormUtils } from "../utils"
import TaskForm from "./TaskForm"
import withFormSubmitter from "./withFormSubmitter"

const EditTaskForm = props => (
  <TaskForm {...props} taskFormType={taskFormTypes.edit} />
)

export default withRouter(
  withFormSubmitter({
    getUrl: props => `${config.FLOTILLA_API}/task/${props.definitionId}`,
    httpMethod: "PUT",
    headers: { "content-type": "application/json" },
    transformFormValues: taskFormUtils.transformFormValues,
    onSuccess: (props, res) => {
      // Go to task definition view.
      props.history.push(`/tasks/${res.definition_id}`)

      // Force refetch.
      props.fetch(props.definitionId)
    },
    onFailure: (props, err) => {
      // console.error(err)
    },
  })(EditTaskForm)
)
