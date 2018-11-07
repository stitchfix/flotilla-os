import React from "react"
import { withRouter } from "react-router-dom"
import TaskForm from "./TaskForm"
import withFormSubmitter from "./withFormSubmitter"
import taskFormTypes from "../constants/taskFormTypes"
import taskFormUtils from "../utils/taskFormUtils"
import config from "../config"

export const CopyTaskForm = props => (
  <TaskForm {...props} taskFormType={taskFormTypes.copy} />
)

export default withRouter(
  withFormSubmitter({
    getUrl: () => `${config.FLOTILLA_API}/task`,
    httpMethod: "POST",
    headers: { "content-type": "application/json" },
    transformFormValues: taskFormUtils.transformFormValues,
    onSuccess: (props, res) => {
      props.history.push(`/tasks/${res.definition_id}`)
    },
  })(CopyTaskForm)
)
