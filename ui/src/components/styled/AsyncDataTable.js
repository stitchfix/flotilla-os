import React from "react"
import styled from "styled-components"
import {
  SPACING_PX,
  NAVIGATION_HEIGHT_PX,
  BREAKPOINTS_PX,
  DETAIL_VIEW_SIDEBAR_WIDTH_PX,
} from "../../constants/styles"

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
