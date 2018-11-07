import React from "react"
import { NavLink, Link, withRouter } from "react-router-dom"
import { get } from "lodash"
import Topbar from "./Topbar"
import Favicon from "../assets/favicon.png"

const FlotillaTopbar = props => {
  // Don't render topbar in <RunMiniView>
  if (get(props, "location.pathname", "").endsWith("/mini")) {
    return <span />
  }

  return (
    <Topbar>
      <div className="pl-topbar-section">
        <Link to="/" className="pl-topbar-app-name">
          <img src={Favicon} alt="stitchfix-logo" className="topbar-logo" />
          <div>FLOTILLA</div>
        </Link>
        <div className="pl-topbar-nav-link-group">
          <NavLink className="pl-topbar-nav-link" to="/tasks">
            Tasks
          </NavLink>
          <NavLink className="pl-topbar-nav-link" to="/runs">
            Runs
          </NavLink>
        </div>
      </div>
    </Topbar>
  )
}

export default withRouter(FlotillaTopbar)
