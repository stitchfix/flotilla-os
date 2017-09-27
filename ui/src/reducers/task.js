import { ActionTypes } from '../constants/'

const initialState = {
  isFetching: false,
  hasError: false,
  task: {},
  runConfig: {},
  _isDeleting: false,
  error: undefined
}

export default function task(state = initialState, action) {
  switch (action.type) {
    case ActionTypes.REQUEST_TASK:
      return ({
        ...state,
        isFetching: true,
        hasError: false,
      })
    case ActionTypes.RECEIVE_TASK:
      return ({
        ...state,
        isFetching: false,
        hasError: false,
        task: action.payload.task
      })
    case ActionTypes.RECEIVE_TASK_ERROR:
      return ({
        ...state,
        hasError: true,
        error: action.payload.error,
        isFetching: false,
      })
    case ActionTypes.RESET_TASK:
      return initialState
    case ActionTypes.RECEIVE_LOCAL_RUN_CONFIG:
      return ({
        ...state,
        runConfig: action.payload.config
      })
    case ActionTypes.REQUEST_DELETE_TASK:
      return ({
        ...state,
        _isDeleting: true
      })
    case ActionTypes.DELETE_TASK_SUCCESS:
      return ({
        ...state,
        _isDeleting: false
      })
    default:
      return state
  }
}
