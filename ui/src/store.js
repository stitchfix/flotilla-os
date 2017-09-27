import { createStore, applyMiddleware } from 'redux'
import thunkMiddleware from 'redux-thunk'
import { routerMiddleware } from 'react-router-redux'
import { hashHistory } from 'react-router'
import createLogger from 'redux-logger'
import reducers from './reducers/'

export default function store() {
  const reduxRouterMiddleware = routerMiddleware(hashHistory)
  let createStoreWithMiddleware
  // Don't log in production.
  if (process.env.NODE_ENV === 'production') {
    createStoreWithMiddleware = applyMiddleware(
      thunkMiddleware,
      reduxRouterMiddleware
    )(createStore)
  } else {
    const logger = createLogger({ collapsed: true })
    createStoreWithMiddleware = applyMiddleware(
      thunkMiddleware,
      reduxRouterMiddleware,
      logger
    )(createStore)
  }

  const _store = createStoreWithMiddleware(reducers)
  return _store
}
