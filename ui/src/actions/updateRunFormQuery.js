// Used to sync run form inputs w/ url query
import { replace } from 'react-router-redux'
import { isEqual } from 'lodash'
import { QueryUpdateTypes } from '../constants/'

// Note: not an actual reducer, just copies the reducer signature.
const queryReducer = (state, action) => {
  if (action.type === QueryUpdateTypes.SHALLOW) {
    return {
      ...state,
      [action.payload.key]: action.payload.value
    }
  } else {
    const nestedState = Array.isArray(state[action.payload.key]) ?
      [...state[action.payload.key]] :
      [state[action.payload.key]]

    if (action.type === QueryUpdateTypes.NESTED_CREATE) {
      return {
        ...state,
        [action.payload.key]: [
          ...nestedState,
          action.payload.value
        ]
      }
    } else if (action.type === QueryUpdateTypes.NESTED_UPDATE) {
      return {
        ...state,
        [action.payload.key]: [
          ...nestedState.slice(0, action.payload.index),
          action.payload.value,
          ...nestedState.slice(action.payload.index + 1),
        ]
      }
    } else if (action.type === QueryUpdateTypes.NESTED_REMOVE) {
      return {
        ...state,
        [action.payload.key]: [
          ...nestedState.slice(0, action.payload.index),
          ...nestedState.slice(action.payload.index + 1),
        ]
      }
    }
  }
}

export default function updateRunFormQuery({ key, value, index, updateType }) {
  return (dispatch, getState) => {
    const { pathname, query } = getState().routing.locationBeforeTransitions
    const nextQuery = queryReducer(query, {
      type: updateType,
      payload: { key, value, index }
    })
    if (!isEqual(query, nextQuery)) {
      dispatch(replace({ pathname, query: nextQuery }))
    }
  }
}
