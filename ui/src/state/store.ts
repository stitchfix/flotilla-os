import {
  configureStore,
  Action,
  combineReducers,
  getDefaultMiddleware,
} from "@reduxjs/toolkit"
import { ThunkAction } from "redux-thunk"
import { createLogger } from "redux-logger"
import settings from "./settings"
import runView from "./runView"

const middleware = [...getDefaultMiddleware()]

// Only use redux-logger in non-production.
if (process.env.NODE_ENV !== "production") {
  const logger = createLogger({
    collapsed: true,
    timestamp: false,
  })

  middleware.push(logger)
}

const rootReducer = combineReducers({
  settings,
  runView,
})

const store = configureStore({
  reducer: rootReducer,
  middleware,
})

export type RootState = ReturnType<typeof rootReducer>
export type AppDispatch = typeof store.dispatch
export type AppThunk = ThunkAction<void, RootState, null, Action<string>>
export default store
