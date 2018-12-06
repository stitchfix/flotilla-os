import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import colors from "../../helpers/colors"
import {
  NAVIGATION_HEIGHT_PX,
  SPACING_PX,
  VIEW_HEADER_HEIGHT_PX,
  DEFAULT_FONT_COLOR,
  Z_INDICES,
  DEFAULT_BORDER,
} from "../../helpers/styles"

const ViewHeaderContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  width: 100vw;
  position: fixed;
  top: ${NAVIGATION_HEIGHT_PX}px;
  left: 0;
  right: 0;
  height: ${VIEW_HEADER_HEIGHT_PX}px;
  background: ${colors.black[0]};
  color: ${DEFAULT_FONT_COLOR};
  z-index: ${Z_INDICES.VIEW_HEADER};
  box-shadow: 0px 2px 8px 1px rgba(0, 0, 0, 0.05);
  border-bottom: ${DEFAULT_BORDER};
`

const ViewHeaderInner = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0 ${SPACING_PX}px;
`

const ViewHeader = ({ children, title, actions }) => (
  <ViewHeaderContainer>
    <ViewHeaderInner>
      {!!title && <h3>{title}</h3>}
      {!!actions && actions}
    </ViewHeaderInner>
  </ViewHeaderContainer>
)

ViewHeader.displayName = "ViewHeader"
ViewHeader.propTypes = {
  actions: PropTypes.node,
  children: PropTypes.node,
  title: PropTypes.node,
}

export default ViewHeader
