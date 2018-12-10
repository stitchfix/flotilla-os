import * as React from "react"
import styled from "styled-components"
import { NAVIGATION_HEIGHT_PX } from "../../helpers/styles"

const ViewContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  width: 100%;
  margin-top: ${NAVIGATION_HEIGHT_PX}px;
`

const ViewInner = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  width: 100%;
`

const View: React.SFC<{}> = ({ children }) => (
  <ViewContainer>
    <ViewInner>{children}</ViewInner>
  </ViewContainer>
)

View.displayName = "View"

export default View
