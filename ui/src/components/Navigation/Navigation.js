import React, { Fragment } from "react"
import PropTypes from "prop-types"
import { NavLink } from "react-router-dom"
import { ChevronRight } from "react-feather"
import { get, isEmpty } from "lodash"
import styled from "styled-components"
import Favicon from "../../assets/favicon.png"
import colors from "../../constants/colors"
import Button from "../styled/Button"
import ButtonLink from "../styled/ButtonLink"
import ButtonGroup from "../styled/ButtonGroup"
import {
  NAVIGATION_HEIGHT_PX,
  SPACING_PX,
  Z_INDICES,
  DEFAULT_BORDER,
  DEFAULT_FONT_COLOR,
} from "../../constants/styles"

const NAVIGATION_BORDER = `1px solid ${colors.black[4]}`
const NAVIGATION_EL_SPACING_PX = SPACING_PX * 2

const NavigationContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  width: 100vw;
  background: ${colors.black[0]};
  height: ${NAVIGATION_HEIGHT_PX}px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: ${Z_INDICES.NAVIGATION};
  border-bottom: ${DEFAULT_BORDER};
`

const NavigationInner = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0 ${SPACING_PX}px;
  position: relative;
  height: 100%;
`

const NavigationSection = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  height: 100%;

  & > * {
    ${({ position }) => {
      if (position === "right") {
        return `margin-left: ${NAVIGATION_EL_SPACING_PX}px;`
      }

      return `margin-right: ${NAVIGATION_EL_SPACING_PX}px;`
    }};
  }
`

const NavigationLink = styled(NavLink)`
  color: ${DEFAULT_FONT_COLOR};

  &.active,
  &:hover {
    color: ${colors.blue[0]};
  }
`

const NavigationBreadcrumbs = styled.div`
  height: 100%;
  border-left: ${NAVIGATION_BORDER};
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  padding-left: ${SPACING_PX}px;
`
const NavigationBreadcrumb = styled(NavLink)``
const NavigationBreadcrumbArrow = styled(ChevronRight)`
  margin: 0 6px;
`

const NavigationLogo = styled.img`
  width: 32px;
  height: 32px;
`

const Navigation = ({ breadcrumbs, actions }) => (
  <NavigationContainer>
    <NavigationInner>
      <NavigationSection position="left">
        <NavigationLink to="/">
          <NavigationLogo src={Favicon} alt="stitchfix-logo" />
        </NavigationLink>
        <NavigationLink to="/tasks">Tasks</NavigationLink>
        <NavigationLink to="/runs">Runs</NavigationLink>
        <NavigationBreadcrumbs>
          {!isEmpty(breadcrumbs) &&
            breadcrumbs.map((b, i) => (
              <Fragment key={i}>
                <NavigationBreadcrumb to={b.href}>
                  {b.text}
                </NavigationBreadcrumb>
                {i !== breadcrumbs.length - 1 && (
                  <NavigationBreadcrumbArrow size="14" />
                )}
              </Fragment>
            ))}
        </NavigationBreadcrumbs>
      </NavigationSection>
      <NavigationSection position="right">
        <ButtonGroup>
          {!isEmpty(actions) &&
            actions.map((a, i) => {
              if (a.isLink === true) {
                return (
                  <ButtonLink
                    key={i}
                    to={a.href}
                    {...get(a, "buttonProps", {})}
                  >
                    {a.text}
                  </ButtonLink>
                )
              }

              return (
                <Button key={i} {...get(a, "buttonProps", {})}>
                  {a.text}
                </Button>
              )
            })}
        </ButtonGroup>
      </NavigationSection>
    </NavigationInner>
  </NavigationContainer>
)

Navigation.displayName = "Navigation"
Navigation.propTypes = {
  actions: PropTypes.arrayOf(
    PropTypes.shape({
      isLink: PropTypes.bool.isRequired,
      text: PropTypes.string.isRequired,
      href: PropTypes.oneOfType([PropTypes.string, PropTypes.object]),
      buttonProps: PropTypes.object,
    })
  ),
  breadcrumbs: PropTypes.arrayOf(
    PropTypes.shape({
      href: PropTypes.string.isRequired,
      text: PropTypes.string.isRequired,
    })
  ),
  children: PropTypes.node,
}

Navigation.defaultProps = {
  actions: [],
  breadcrumbs: [],
}

export default Navigation
