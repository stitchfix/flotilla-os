import React, { SFC } from "react"
import { BrowserRouter, Switch, Route, Redirect } from "react-router-dom"

const App: SFC<{}> = () => (
  <BrowserRouter>
    <Switch>
      <Route exact path="/" component={() => <div>hi</div>} />
    </Switch>
  </BrowserRouter>
)

export default App
