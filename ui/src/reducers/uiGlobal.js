import { ActionTypes } from '../constants/'

const initialState = {
  modal: {
    visible: false,
    modal: null
  }
}

export default function uiGlobal(state = initialState, action) {
  switch (action.type) {
    case ActionTypes.RENDER_MODAL:
      return ({
        ...state,
        modal: {
          ...state.modal,
          visible: true,
          modal: action.payload.modal
        }
      })
    case ActionTypes.UNRENDER_MODAL:
      return ({
        ...state,
        modal: {
          ...state.modal,
          visible: false,
          modal: null
        }
      })
    default:
      return state
  }
}
