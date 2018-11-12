import React from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import styled from "styled-components"
import Favicon from "../../assets/favicon.png"
import colors from "../../constants/colors"
import {
  TOPBAR_HEIGHT_PX,
  SPACING_PX,
  Z_INDICES,
  DEFAULT_BORDER,
  DEFAULT_FONT_COLOR,
} from "../../constants/styles"

const NAVIGATION_LINK_BORDER = `1px solid ${colors.black[4]}`

const NavigationContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  width: 100vw;
  background: ${colors.black[2]};
  height: ${TOPBAR_HEIGHT_PX}px;
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
  color: ${DEFAULT_FONT_COLOR};
  height: 100%;
`

const NavigationTitle = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  border: none;
  color: ${DEFAULT_FONT_COLOR};
  font-size: 1.05rem;
  font-weight: 300;
  height: 100%;
  letter-spacing: 0.04rem;
  margin-right: ${SPACING_PX}px;
  padding: 0;
`

const NavigationLinkGroup = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  height: 100%;

  & > * {
    border-left: ${NAVIGATION_LINK_BORDER};
    &:last-child {
      border-right: ${NAVIGATION_LINK_BORDER};
    }
  }
`

const NavigationLink = styled(Link)`
  display: flex;
  align-items: center;
  color: ${DEFAULT_FONT_COLOR};
  border-left: ${NAVIGATION_LINK_BORDER};
  padding: 0 ${SPACING_PX}px;
  height: 100%;

  &:last-child {
    border-right: ${NAVIGATION_LINK_BORDER};
  }

  &.active,
  &:hover {
    color: ${colors.blue[0]};
    background: ${colors.black[1]};
  }
`

const NavigationLogo = styled.img`
  width: 32px;
  height: 32px;
  border-radius: 6px;
  margin-right: 8px;
`

const Navigation = () => (
  <NavigationContainer>
    <NavigationInner>
      <NavigationSection>
        <NavigationTitle>
          <NavigationLogo src={Favicon} alt="stitchfix-logo" />
          <div>Flotilla</div>
        </NavigationTitle>
        <NavigationLinkGroup>
          <NavigationLink to="/tasks">Tasks</NavigationLink>
          <NavigationLink to="/runs">Runs</NavigationLink>
        </NavigationLinkGroup>
      </NavigationSection>
    </NavigationInner>
  </NavigationContainer>
)

Navigation.displayName = "Navigation"
Navigation.propTypes = {
  children: PropTypes.node,
}

export default Navigation
