import { ActionTypes, getApiRoot } from '../constants/'
import { checkStatus } from '../utils/'

const requestDeleteTask = () => ({
  type: ActionTypes.REQUEST_DELETE_TASK
})

const deleteTaskSuccess = () => ({
  type: ActionTypes.DELETE_TASK_SUCCESS
})

const deleteTaskError = error => ({
  type: ActionTypes.DELETE_TASK_ERROR
})

export default function deleteTask({ taskID }, cb) {
  return (dispatch) => {
    dispatch(requestDeleteTask())

    fetch(`${getApiRoot()}/task/${taskID}`, {
      method: 'DELETE'
    })
      .then(checkStatus)
      .then(res => res.json())
      .then((res) => {
        dispatch(deleteTaskSuccess())
        cb()
      })
      .catch((err) => {
        dispatch(deleteTaskError(err))
        console.error(err)
      })
  }
}
