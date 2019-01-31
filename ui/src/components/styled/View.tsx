import * as React from "react"
import styled from "styled-components"
import { NAVIGATION_HEIGHT_PX } from "../../helpers/styles"

const ViewContainer = styled.div`
  height: calc(100vh - ${NAVIGATION_HEIGHT_PX}px);
  overflow-y: hidden;
  margin-top: ${NAVIGATION_HEIGHT_PX}px;
`

const View: React.SFC<{}> = ({ children }) => (
  <ViewContainer>{children}</ViewContainer>
)

View.displayName = "View"

export default View
