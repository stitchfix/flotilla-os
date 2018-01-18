import { get, has, isEmpty } from "lodash"
import config from "../config"
import { taskFormTypes } from "../constants/"

export const joinImage = (image, tag) =>
  `${config.DOCKER_REPOSITORY_HOST}/${image}:${tag}`

export const splitImage = str => {
  const split = str.split("/")[1].split(":")
  return {
    image: split[0],
    tag: split[1],
  }
}

// Set initial values and request image tags (for non-create forms)
export const mapStateToProps = (state, ownProps) => {
  const ret = {
    selectOptionsRequestInFlight: state.selectOpts.inFlight,
    imageOptions: get(state, "selectOpts.image", []),
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
    const vals = ["group_name", "command", "memory", "env", "tags"]

    vals.forEach(val => {
      if (has(ownProps.data, val)) {
        initialValues[val] = ownProps.data[val]
      }
    })

    if (has(ownProps.data, "image")) {
      const split = splitImage(ownProps.data.image)
      initialValues.image = split.image
      initialValues.image_tag = split.tag
    }

    if (!isEmpty(initialValues)) {
      ret.initialValues = initialValues
    }
  }

  return ret
}

// Concat image and image tag; add env and tags if necessary.
export const transformFormValues = values => {
  const body = {
    alias: values.alias,
    command: values.command,
    group_name: values.group_name,
    // @TODO: get docker repository url from config file.
    image: joinImage(values.image, values.image_tag),
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

  if (!values.alias) errors.alias = "You must enter an alias."
  if (!values.group_name) errors.group_name = "You must select a group name."
  if (!values.image) errors.image = "You must select an image."
  if (!values.image_tag) errors.image_tag = "You must select an image tag."
  if (!values.command) errors.command = "You must enter a command."
  if (!values.memory) errors.memory = "You must enter a memory value."

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
