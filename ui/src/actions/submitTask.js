import { push } from 'react-router-redux'
import config from '../config'
import { getApiRoot, TaskFormTypes } from '../constants/'
import { fetchTask } from './'

// This is used for both creating new and updating existing tasks.
export default function submitTask({ values, formType, id }) {
  return (dispatch) => {
    const url = formType === TaskFormTypes.edit ?
      `${getApiRoot()}/task/${id}` :
      `${getApiRoot()}/task`
    const method = formType === TaskFormTypes.edit ? 'put' : 'post'
    const body = {
      alias: values.alias,
      command: values.command,
      group_name: values.group,
      // [OS] you may need to modify this string.
      image: `${config.DOCKER_REPOSITORY_HOST}/${values.image}:${values.imageTag}`,
      memory: +values.memory,
    }

    if (!!values.env && Array.isArray(values.env)) {
      // TODO: add validation before pushing to prod.
      body.env = values.env
    }

    if (!!values.tags) {
      if (Array.isArray(values.tags)) {
        body.tags = values.tags
      } else if (typeof values.tags === 'string') {
        body.tags = values.tags.split(',')
      }
    }

    fetch(url, {
      method,
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(body)
    })
      .then(res => res.json())
      .then((res) => {
        // Need to manually refetch the task definition, as the
        // componentDidMount method in TaskContainer has already
        // been called.
        if (formType === TaskFormTypes.edit) {
          dispatch(fetchTask({ id: res.definition_id }))
        }
        dispatch(push(`/tasks/${res.definition_id}`))
      })
  }
}
