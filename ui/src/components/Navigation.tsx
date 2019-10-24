import * as React from "react"
import { Link, NavLink } from "react-router-dom"
import {
  ButtonGroup,
  Navbar,
  NavbarDivider,
  NavbarGroup,
  Alignment,
  Classes,
} from "@blueprintjs/core"

const Navigation: React.FunctionComponent = () => (
  <Navbar fixedToTop className="bp3-dark">
    <NavbarGroup align={Alignment.LEFT}>
      <Link to="/tasks" className="bp3-button bp3-minimal">
        Flotilla
      </Link>
      <NavbarDivider />
      <ButtonGroup className={Classes.MINIMAL}>
        <NavLink exact to="/tasks" className={Classes.BUTTON}>
          Tasks
        </NavLink>
        <NavLink exact to="/runs" className={Classes.BUTTON}>
          Runs
        </NavLink>
      </ButtonGroup>
    </NavbarGroup>
    <NavbarGroup align={Alignment.RIGHT}>
      <a
        href="https://github.com/stitchfix/flotilla-os"
        className={Classes.BUTTON}
      >
        Github
      </a>
    </NavbarGroup>
  </Navbar>
)

export default Navigation
