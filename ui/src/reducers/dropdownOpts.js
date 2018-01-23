import { ActionTypes } from '../constants/'

const initialState = {
  cluster: [],
  group: [],
  image: [],
  tag: [],
  hasError: false,
  isFetching: false,
}

export default function dropdownOpts(state = initialState, action) {
  switch (action.type) {
    case ActionTypes.REQUEST_DROPDOWN_OPTS:
      return ({
        ...state,
        isFetching: true,
        hasError: false,
      })
    case ActionTypes.RECEIVE_DROPDOWN_OPTS:
      return ({
        ...state,
        isFetching: false,
        hasError: false,
        cluster: action.payload.cluster,
        group: action.payload.group,
        image: action.payload.image,
        tag: action.payload.tag,
      })
    case ActionTypes.RECEIVE_DROPDOWN_OPTS_ERROR:
      return ({
        ...state,
        hasError: true,
        error: action.payload.error
      })
    default:
      return state
  }
}
