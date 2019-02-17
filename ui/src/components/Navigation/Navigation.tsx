import * as React from "react"
import { get, isEmpty } from "lodash"
import { ChevronRight } from "react-feather"
import Button from "../styled/Button"
import ButtonLink from "../styled/ButtonLink"
import ButtonGroup from "../styled/ButtonGroup"
import { IFlotillaUINavigationLink, IFlotillaUIBreadcrumb } from "../../types"
import {
  NavigationContainer,
  NavigationInner,
  NavigationSection,
  NavigationLink,
  NavigationBreadcrumbs,
  NavigationBreadcrumb,
} from "../styled/Navigation"

interface INavigationProps {
  actions: IFlotillaUINavigationLink[]
  breadcrumbs: IFlotillaUIBreadcrumb[]
}

class Navigation extends React.PureComponent<INavigationProps> {
  static displayName = "Navigation"
  static defaultProps: Partial<INavigationProps> = {
    actions: [],
    breadcrumbs: [],
  }
  render() {
    const { actions, breadcrumbs } = this.props
    return (
      <NavigationContainer>
        <NavigationInner>
          <NavigationSection position="left">
            <NavigationLink to="/">
              {/* <NavigationLogo src={Favicon} alt="flotilla-logo" /> */}
            </NavigationLink>
            <NavigationLink to="/tasks">Tasks</NavigationLink>
            <NavigationLink to="/runs">Runs</NavigationLink>
            <NavigationBreadcrumbs>
              {!isEmpty(breadcrumbs) &&
                breadcrumbs.map((b, i) => (
                  <React.Fragment key={i}>
                    <NavigationBreadcrumb to={b.href}>
                      {b.text}
                    </NavigationBreadcrumb>
                    {i !== breadcrumbs.length - 1 && (
                      <ChevronRight size={14} style={{ margin: "0 6px" }} />
                    )}
                  </React.Fragment>
                ))}
            </NavigationBreadcrumbs>
          </NavigationSection>
          <NavigationSection position="right">
            <ButtonGroup>
              {!!actions &&
                !isEmpty(actions) &&
                actions.map((a, i) => {
                  const buttonProps: any = a.buttonProps
                  if (a.isLink === true && !!a.href) {
                    return (
                      <ButtonLink key={i} to={a.href} {...buttonProps}>
                        {a.text}
                      </ButtonLink>
                    )
                  }

                  return (
                    <Button key={i} {...buttonProps}>
                      {a.text}
                    </Button>
                  )
                })}
            </ButtonGroup>
          </NavigationSection>
        </NavigationInner>
      </NavigationContainer>
    )
  }
}

export default Navigation
