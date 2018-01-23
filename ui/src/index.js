import React from 'react'
import { render } from 'react-dom'
import { Provider } from 'react-redux'
import { Router, Route, hashHistory, Redirect } from 'react-router'
import { syncHistoryWithStore } from 'react-router-redux'
import {
  App,
  MainContainer,
  RunContainer,
  RunFormContainer,
  Runs,
  Task,
  TaskContainer,
  TaskFormContainer,
  Tasks,
  Run,
  RunMini,
} from './containers/'
import './styles/index.scss'
import store from './store'

const _store = store()
const history = syncHistoryWithStore(hashHistory, _store)

render(
  <Provider store={_store}>
    <Router history={history}>
      <Route component={App}>
        <Route path="create-task" component={TaskFormContainer} />
        <Route component={MainContainer}>
          <Route path="tasks" component={Tasks} />
          <Route path="runs" component={Runs} />
        </Route>
        <Route component={TaskContainer}>
          <Route path="tasks/:taskID" component={Task} />
          <Route path="tasks/:taskID/edit" component={TaskFormContainer} />
          <Route path="tasks/:taskID/copy" component={TaskFormContainer} />
          <Route path="tasks/:taskID/run" component={RunFormContainer} />
        </Route>
        <Route component={RunContainer}>
          <Route path="runs/:runID" component={Run} />
          <Route path="runs/:runID/mini" component={RunMini} />
        </Route>
      </Route>
      <Redirect from="/" to="/tasks" />
    </Router>
  </Provider>,
  document.getElementById('root')
)
