import { actionTypes } from "../constants/"

const initialState = {
  cluster: [],
  group: [],
  tag: [],
  error: false,
  inFlight: false,
}

export default function selectOpts(state = initialState, action) {
  switch (action.type) {
    case actionTypes.REQUEST_SELECT_OPTS:
      return {
        ...state,
        inFlight: true,
        error: false,
      }
    case actionTypes.RECEIVE_SELECT_OPTS:
      return {
        ...state,
        inFlight: false,
        error: false,
        cluster: action.payload.cluster,
        group: action.payload.group,
        tag: action.payload.tag,
      }
    case actionTypes.RECEIVE_SELECT_OPTS_ERROR:
      return {
        ...state,
        error: true,
        error: action.payload,
      }
    default:
      return state
  }
}
