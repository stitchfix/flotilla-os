/// <reference path="../globals.d.ts" />

import React, { Fragment } from "react"
import { render } from "react-dom"
import App from "./components/App/App"
import GlobalStyle from "./components/styled/GlobalStyle"

render(
  <Fragment>
    <GlobalStyle />
    <App />
  </Fragment>,
  document.querySelector("#root")
)
