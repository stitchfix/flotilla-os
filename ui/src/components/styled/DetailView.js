import s from "styled-components"
import {
  DEFAULT_BORDER,
  DETAIL_VIEW_SIDEBAR_WIDTH_PX,
  NAVIGATION_HEIGHT_PX,
} from "../../constants/styles"
import colors from "../../constants/colors"

export const DetailViewContainer = s.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: flex-start;
  width: 100%;
`

export const DetailViewContent = s.div`
  margin-left: ${DETAIL_VIEW_SIDEBAR_WIDTH_PX + 24}px;
  flex: 1;
  height: calc(100vh - ${NAVIGATION_HEIGHT_PX}px);
`

export const DetailViewSidebar = s.div`
  position: fixed;
  top: ${NAVIGATION_HEIGHT_PX}px;
  bottom: 0;
  left: 0;
  width: ${DETAIL_VIEW_SIDEBAR_WIDTH_PX}px;
  border-right: ${DEFAULT_BORDER};
  background: ${colors.black[0]};
  overflow-y: scroll;
`
