import styled, { css } from "styled-components"
import colors from "../../helpers/colors"
import { DEFAULT_BORDER } from "../../helpers/styles"

export const Table = styled.table`
  background: ${colors.black[0]};
  display: flex;
  flex-flow: column nowrap;
  overflow: hidden;
  width: 100%;
`

export const TableRow = styled.tr`
  align-items: center;
  background: inherit;
  border-bottom: 1px solid ${colors.black[3]};
  display: flex;
  flex-flow: row nowrap;
  height: 48px;
  justify-content: flex-start;
  max-width: 100%;
  width: 100%;

  &:hover {
    background: ${colors.black[2]};
  }
`

const cellStyles = css`
  align-items: center;
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  min-width: 0;
  overflow: hidden;
  padding: 0 8px;
  text-overflow: ellipsis;
  white-space: nowrap;
  height: 100%;
  /* border-left: ${DEFAULT_BORDER}; */
`

export const TableCell = styled.td`
  ${cellStyles};
  flex: ${({ width }: { width?: number }) => (!!width ? width : 1)};
`

export const TableHeaderSortIcon = styled.div`
  margin-left: 4px;
  font-size: 0.6rem;
`

export const TableHeaderCell = styled.th`
  ${cellStyles};
  font-size: 0.9rem;
  text-transform: uppercase;
  font-weight: 500;
  border-top: none;
  flex: ${({ width }: { width?: number }) => (!!width ? width : 1)};
`

export const TableHeaderCellSortable =
  styled(TableHeaderCell) <
  { isActive: boolean, width: number } >
  `
  cursor: pointer;
  color: ${({ isActive }: { isActive?: boolean }) =>
    isActive ? colors.blue[0] : colors.gray[4]};
  background: ${({ isActive }: { isActive?: boolean }) =>
    isActive ? colors.black[1] : colors.black[0]};

  &:hover {
    color: ${colors.blue[0]};
    background: ${colors.black[1]};
  }
`
