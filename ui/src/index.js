import React from "react"
import { render } from "react-dom"
import { Provider } from "react-redux"
import "./styles/index.scss"
import Store from "./store"
import App from "./components/App"
import fetchSelectOpts from "./actions/fetchSelectOpts"

const store = Store()

// Dispatch global select options (group, tags, etc.)
store.dispatch(fetchSelectOpts())

render(
  <Provider store={store}>
    <App />
  </Provider>,
  document.getElementById("root")
)
