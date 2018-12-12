import { NavLink } from "react-router-dom"
import styled from "styled-components"
import colors from "../../helpers/colors"
import {
  NAVIGATION_HEIGHT_PX,
  SPACING_PX,
  Z_INDICES,
  DEFAULT_BORDER,
  DEFAULT_FONT_COLOR,
} from "../../helpers/styles"

const NAVIGATION_BORDER = `1px solid ${colors.black[4]}`
const NAVIGATION_EL_SPACING_PX = SPACING_PX * 2

export const NavigationContainer = styled.div`
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

export const NavigationInner = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0 ${SPACING_PX}px;
  position: relative;
  height: 100%;
`

export const NavigationSection = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  height: 100%;

  & > * {
    ${({ position }: { position: "left" | "right" }) => {
      if (position === "right") {
        return `margin-left: ${NAVIGATION_EL_SPACING_PX}px;`
      }

      return `margin-right: ${NAVIGATION_EL_SPACING_PX}px;`
    }};
  }
`

export const NavigationLink = styled(NavLink)`
  color: ${DEFAULT_FONT_COLOR};

  &.active,
  &:hover {
    color: ${colors.blue[0]};
  }
`

export const NavigationBreadcrumbs = styled.div`
  height: 100%;
  border-left: ${NAVIGATION_BORDER};
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: center;
  padding-left: ${SPACING_PX}px;
`
export const NavigationBreadcrumb = styled(NavLink)``
export const NavigationLogo = styled.img`
  width: 32px;
  height: 32px;
`
