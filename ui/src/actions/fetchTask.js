import { ActionTypes, getApiRoot } from '../constants/'
import { checkStatus } from '../utils/'

const requestTask = () => ({ type: ActionTypes.REQUEST_TASK })
const receiveTask = task => ({
  type: ActionTypes.RECEIVE_TASK,
  payload: { task }
})
const receiveTaskError = error => ({
  type: ActionTypes.RECEIVE_TASK_ERROR,
  payload: { error }
})


export default function fetchTask({ id }) {
  return (dispatch) => {
    dispatch(requestTask())

    if (!!id) {
      fetch(`${getApiRoot()}/task/${id}`)
        .then(checkStatus)
        .then(res => res.json())
        .then((res) => { dispatch(receiveTask(res)) })
        .catch((error) => { dispatch(receiveTaskError(error)) })
    } else {
      console.error(`No \`id\` value was passed to the fetchTask action.`)
    }
  }
}
