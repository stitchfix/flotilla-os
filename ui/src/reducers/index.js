import { combineReducers } from "redux"
import { reducer as form } from "redux-form"
import { modalReducer, popupReducer } from "platforma"
import selectOpts from "./selectOpts"

export default combineReducers({
  selectOpts,
  modal: modalReducer,
  popup: popupReducer,
  form,
})
