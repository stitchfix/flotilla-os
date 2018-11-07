import { combineReducers } from "redux"
import { reducer as form } from "redux-form"
import modalReducer from "./modalReducer"
import popupReducer from "./popupReducer"
import selectOpts from "./selectOpts"

export default combineReducers({
  selectOpts,
  modal: modalReducer,
  popup: popupReducer,
  form,
})
