import { get, has, isEmpty } from "lodash"
import config from "../config"
import { taskFormTypes } from "../constants/"

// Set initial values (for non-create forms).
export const mapStateToProps = (state, ownProps) => {
  const ret = {
    selectOptionsRequestInFlight: state.selectOpts.inFlight,
    groupOptions: get(state, "selectOpts.group", []),
    tagOptions: get(state, "selectOpts.tag", []),
  }

  // Add initial values.
  if (
    (ownProps.taskFormType === taskFormTypes.edit ||
      ownProps.taskFormType === taskFormTypes.copy) &&
    !isEmpty(ownProps.data)
  ) {
    const initialValues = {}
    const vals = ["group_name", "command", "memory", "env", "tags", "image"]

    vals.forEach(val => {
      if (has(ownProps.data, val)) {
        initialValues[val] = ownProps.data[val]
      }
    })

    if (!isEmpty(initialValues)) {
      ret.initialValues = initialValues
    }
  }

  return ret
}

export const transformFormValues = values => {
  const body = {
    alias: values.alias,
    command: values.command,
    group_name: values.group_name,
    image: values.image,
    memory: +values.memory,
  }

  if (!!values.env && Array.isArray(values.env)) {
    // TODO: add validation before pushing to prod.
    body.env = values.env
  }

  if (!!values.tags) {
    if (Array.isArray(values.tags)) {
      body.tags = values.tags
    } else if (typeof values.tags === "string") {
      body.tags = values.tags.split(",")
    }
  }

  return body
}

// Validation
export const validate = values => {
  const defaultErrMessage = "This is a required field."
  const errors = {}

  if (!values.alias) errors.alias = defaultErrMessage
  if (!values.group_name) errors.group_name = defaultErrMessage
  if (!values.image) errors.image = defaultErrMessage
  if (!values.command) errors.command = defaultErrMessage
  if (!values.memory) errors.memory = defaultErrMessage

  if (!!values.env && values.env.length > 0) {
    const envErr = []
    values.env.forEach((e, i) => {
      const envvarErr = {}
      if (!!e) {
        if (!e.value) {
          envvarErr.value = defaultErrMessage
          envErr[i] = envvarErr
        }
        if (!e.name) {
          envvarErr.name = defaultErrMessage
          envErr[i] = envvarErr
        }
      }
    })
    if (envErr.length > 0) {
      errors.env = envErr
    }
  }
  return errors
}
