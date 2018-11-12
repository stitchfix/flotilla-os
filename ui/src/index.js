import React, { Fragment } from "react"
import { render } from "react-dom"
import { Provider } from "react-redux"

import Store from "./store"
import App from "./components/App/App"
import fetchSelectOpts from "./actions/fetchSelectOpts"
import GlobalStyle from "./components/styled/GlobalStyle"

const store = Store()

// Dispatch global select options (group, tags, etc.)
store.dispatch(fetchSelectOpts())

render(
  <Fragment>
    <GlobalStyle />
    <Provider store={store}>
      <App />
    </Provider>
  </Fragment>,
  document.getElementById("root")
)
