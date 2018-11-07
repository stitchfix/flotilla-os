import actionTypes from "../constants/actionTypes"

const initialState = {
  popupVisible: false,
  popup: undefined,
}

export default function popup(state = initialState, action) {
  switch (action.type) {
    case actionTypes.RENDER_POPUP:
      return {
        ...state,
        popupVisible: true,
        popup: action.payload,
      }
    case actionTypes.UNRENDER_POPUP:
      return {
        ...state,
        popupVisible: false,
        popup: undefined,
      }
    default:
      return state
  }
}
