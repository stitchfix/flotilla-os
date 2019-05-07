import * as React from "react"
import {
  TableHeaderCell,
  TableHeaderCellSortable,
  TableHeaderSortIcon,
} from "../styled/Table"
import { SortOrders } from "../ListRequest/ListRequest"

export interface IProps {
  isSortable: boolean
  currentSortKey: string | null
  currentSortOrder: SortOrders
  children: React.ReactNode
  sortKey: string
  width: number
  onClick: () => void
}

const DataTableHeader: React.FunctionComponent<IProps> = ({
  isSortable,
  currentSortKey,
  currentSortOrder,
  children,
  sortKey,
  width,
  onClick,
}) => {
  if (isSortable) {
    const isActive = currentSortKey === sortKey

    return (
      <TableHeaderCellSortable
        onClick={onClick}
        width={width}
        isActive={isActive}
      >
        {children}
        {!!isActive && (
          <TableHeaderSortIcon>
            {currentSortOrder === SortOrders.ASC ? "▲" : "▼"}
          </TableHeaderSortIcon>
        )}
      </TableHeaderCellSortable>
    )
  }

  return <TableHeaderCell width={width}>{children}</TableHeaderCell>
}

export default DataTableHeader
