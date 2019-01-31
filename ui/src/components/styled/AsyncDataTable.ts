import styled from "styled-components"
import colors from "../../helpers/colors"
import {
  SPACING_PX,
  NAVIGATION_HEIGHT_PX,
  DETAIL_VIEW_SIDEBAR_WIDTH_PX,
} from "../../helpers/styles"

const ASYNC_DATA_TABLE_FILTERS_WIDTH_PX = 280

export const AsyncDataTableContent = styled.div`
  flex: 1;
  margin-left: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  height: calc(100vh - ${NAVIGATION_HEIGHT_PX}px);
  overflow-y: scroll;
`

export const AsyncDataTableFilters = styled.div`
  bottom: 0;
  left: ${({ isView }: { isView?: boolean }) =>
    isView ? 0 : DETAIL_VIEW_SIDEBAR_WIDTH_PX}px;
  max-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  min-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  overflow-y: scroll;
  padding: ${SPACING_PX}px;
  position: fixed;
  top: ${NAVIGATION_HEIGHT_PX}px;
`

export const AsyncDataTableContainer = styled.div`
  position: relative;
  width: 100%;
`

export const AsyncDataTableLoadingMask = styled.div`
  align-items: center;
  background: ${colors.black[0]}99;
  bottom: 0;
  display: flex;
  flex-flow: row nowrap;
  height: 100%;
  justify-content: center;
  left: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  position: absolute;
  right: 0;
  top: 0;
  width: calc(100% - ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px);
`
