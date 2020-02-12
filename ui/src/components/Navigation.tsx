import * as React from "react"
import {
  Link,
  NavLink,
  withRouter,
  RouteComponentProps,
} from "react-router-dom"
import {
  ButtonGroup,
  Navbar,
  NavbarDivider,
  NavbarGroup,
  Alignment,
  Classes,
  Button,
  Menu,
  MenuItem,
  Popover,
  Tag,
  Intent,
} from "@blueprintjs/core"
import SettingsButton from "./SettingsButton"

const Navigation: React.FunctionComponent<RouteComponentProps> = ({
  location,
}) => {
  let executableDropdownButtonChild: React.ReactNode = "Executable / Tasks"

  if (location.pathname.startsWith("/templates")) {
    executableDropdownButtonChild = (
      <>
        <span style={{ marginRight: 8 }}>Executable / Templates</span>
        <Tag intent={Intent.DANGER}>EXPERIMENTAL</Tag>
      </>
    )
  }

  return (
    <Navbar fixedToTop className="bp3-dark">
      <NavbarGroup align={Alignment.LEFT}>
        <Link to="/tasks" className="bp3-button bp3-minimal">
          Flotilla
        </Link>
        <NavbarDivider />
        <ButtonGroup className={Classes.MINIMAL}>
          <Popover
            content={
              <Menu>
                <NavLink to="/templates" className={Classes.MENU_ITEM}>
                  <span>Templates</span>
                  <Tag intent={Intent.DANGER}>EXPERIMENTAL</Tag>
                </NavLink>
                <NavLink to="/tasks" className={Classes.MENU_ITEM}>
                  Task Definitions
                </NavLink>
              </Menu>
            }
          >
            <Button rightIcon="caret-down">
              {executableDropdownButtonChild}
            </Button>
          </Popover>
          <NavLink exact to="/runs" className={Classes.BUTTON}>
            Runs
          </NavLink>
        </ButtonGroup>
      </NavbarGroup>
      <NavbarGroup align={Alignment.RIGHT}>
        <ButtonGroup>
          <SettingsButton />
          <a
            href="https://github.com/stitchfix/flotilla-os"
            className={Classes.BUTTON}
          >
            Github
          </a>
        </ButtonGroup>
      </NavbarGroup>
    </Navbar>
  )
}

export default withRouter(Navigation)
