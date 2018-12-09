import React, { Fragment } from "react"
import { render } from "react-dom"
import App from "./components/App/App"

render(
  <Fragment>
    <App />
  </Fragment>,
  document.querySelector("#root")
)
