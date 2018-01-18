import { createStore, applyMiddleware, combineReducers } from "redux"
import thunkMiddleware from "redux-thunk"
import { createLogger } from "redux-logger"
import reducers from "./reducers"

export default function configureStore(initialState) {
  const create = window.devToolsExtension
    ? window.devToolsExtension()(createStore)
    : createStore

  let createStoreWithMiddleware

  if (process.env.NODE_ENV === "production") {
    createStoreWithMiddleware = applyMiddleware(thunkMiddleware)(create)
  } else {
    const logger = createLogger({
      collapsed: true,
      timestamp: false,
    })
    createStoreWithMiddleware = applyMiddleware(thunkMiddleware, logger)(create)
  }

  const store = createStoreWithMiddleware(reducers, initialState)

  return store
}
