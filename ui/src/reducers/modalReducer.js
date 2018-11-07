import actionTypes from "../constants/actionTypes"

const initialState = {
  modalVisible: false,
  modal: undefined,
}

export default function modal(state = initialState, action) {
  switch (action.type) {
    case actionTypes.RENDER_MODAL:
      return {
        ...state,
        modalVisible: true,
        modal: action.payload.modal,
      }
    case actionTypes.UNRENDER_MODAL:
      return {
        ...state,
        modalVisible: false,
        modal: undefined,
      }
    default:
      return state
  }
}
