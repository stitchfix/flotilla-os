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
`

export const AsyncDataTableFilters = styled.div`
  padding: ${SPACING_PX}px;
  min-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  max-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  position: fixed;
  top: ${NAVIGATION_HEIGHT_PX}px;
  left: ${({ isView }) => (isView ? 0 : DETAIL_VIEW_SIDEBAR_WIDTH_PX)}px;
  bottom: 0;
  overflow-y: scroll;
`

export const AsyncDataTableContainer = styled.div`
  width: 100%;
  position: relative;
`

export const AsyncDataTableLoadingMask = styled.div`
  width: calc(100% - ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px);
  height: 100%;
  position: absolute;
  left: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  background: ${colors.black[0]}99;
`
