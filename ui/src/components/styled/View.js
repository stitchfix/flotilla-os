import React from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import {
  TOPBAR_HEIGHT_PX,
  SPACING_PX,
  VIEW_HEADER_HEIGHT_PX,
} from "../../constants/styles"

const ViewContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  width: 100vw;
  margin-top: ${TOPBAR_HEIGHT_PX}px;
`

const ViewInner = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  /* margin-top: ${VIEW_HEADER_HEIGHT_PX}px; */
`

const View = ({ children }) => (
  <ViewContainer>
    <ViewInner>{children}</ViewInner>
  </ViewContainer>
)

View.displayName = "View"

View.propTypes = {
  children: PropTypes.node,
}

export default View
