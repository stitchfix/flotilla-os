import React from "react"
import styled from "styled-components"
import {
  SPACING_PX,
  TOPBAR_HEIGHT_PX,
  VIEW_HEADER_HEIGHT_PX,
} from "../../constants/styles"

const ASYNC_DATA_TABLE_FILTERS_WIDTH_PX = 360

export const AsyncDataTableContainer = styled.div`
  width: 100%;
`

export const AsyncDataTableFilters = styled.div`
  padding: ${SPACING_PX}px;
  min-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  position: fixed;
  top: calc(${TOPBAR_HEIGHT_PX}px + ${VIEW_HEADER_HEIGHT_PX}px);
  left: 0;
  bottom: 0;
`

export const AsyncDataTableContent = styled.div`
  flex: 1;
  margin-left: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  overflow-y: scroll;
  height: calc(100vh - ${TOPBAR_HEIGHT_PX}px - ${VIEW_HEADER_HEIGHT_PX}px);
`
