import { combineReducers } from "redux"
import { reducer as form } from "redux-form"
import { modalReducer, popupReducer } from "aa-ui-components"
import selectOpts from "./selectOpts"

export default combineReducers({
  selectOpts,
  modal: modalReducer,
  popup: popupReducer,
  form,
})
