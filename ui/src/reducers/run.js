import { ActionTypes } from '../constants/'

const initialState = {
  infoIsFetching: false,
  infoHasError: false,
  infoError: undefined,
  logsAreFetching: false,
  logsHaveError: false,
  logsError: undefined,
  info: {},
  logs: [],
  lastSeen: undefined,
}

export default function run(state = initialState, action) {
  switch (action.type) {
    case ActionTypes.REQUEST_RUN_INFO:
      return ({
        ...state,
        infoIsFetching: true,
        infoHasError: false,
      })
    case ActionTypes.RECEIVE_RUN_INFO:
      return ({
        ...state,
        infoIsFetching: false,
        infoHasError: false,
        info: action.payload.info
      })
    case ActionTypes.RECEIVE_RUN_INFO_ERROR:
      return ({
        ...state,
        infoIsFetching: false,
        infoHasError: true,
        infoError: action.payload.error
      })
    case ActionTypes.REQUEST_RUN_LOGS:
      return ({
        ...state,
        logsAreFetching: true,
        logsHaveError: false,
      })
    case ActionTypes.RECEIVE_RUN_LOGS:
      return ({
        ...state,
        logsAreFetching: false,
        logsHaveError: false,
        logs: [
          ...state.logs,
          {
            lastSeen: action.payload.lastSeen,
            logs: action.payload.logs.split(/\n/)
          }
        ],
        lastSeen: action.payload.lastSeen,
      })
    case ActionTypes.RECEIVE_RUN_LOGS_ERROR:
      return ({
        ...state,
        logsAreFetching: false,
        logsHaveError: true,
        logsError: action.payload.error
      })
    case ActionTypes.STOP_RUN_INTERVAL:
      return ({
        ...state,
        infoIsFetching: false,
        logsAreFetching: false,
      })
    case ActionTypes.RESET_RUN:
      return initialState
    default:
      return state
  }
}
