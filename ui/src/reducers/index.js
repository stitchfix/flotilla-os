import { combineReducers } from 'redux'
import { routerReducer } from 'react-router-redux'
import { reducer as form } from 'redux-form'
import uiGlobal from './uiGlobal'
import dropdownOpts from './dropdownOpts'
import task from './task'
import run from './run'

export default combineReducers({
  uiGlobal,
  dropdownOpts,
  task,
  run,
  form,
  routing: routerReducer
})
