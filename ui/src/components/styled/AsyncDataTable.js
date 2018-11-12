import styled from "styled-components"
import { SPACING_PX } from "../../constants/styles"

const ASYNC_DATA_TABLE_FILTERS_WIDTH_PX = 360

export const AsyncDataTableContainer = styled.div`
  width: 100%;
  position: relative;
`

export const AsyncDataTableFilters = styled.div`
  padding: ${SPACING_PX}px;
  min-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  max-width: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
`

export const AsyncDataTableContent = styled.div`
  flex: 1;
  margin-left: ${ASYNC_DATA_TABLE_FILTERS_WIDTH_PX}px;
`
