import * as React from "react"
import { Table, TableRow, TableCell } from "../styled/Table"
import DataTableHeader from "./DataTableHeader"
import { SortOrders } from "../ListRequest/ListRequest"
import { flotillaUIRequestStates } from "../../types"
import Loader from "../styled/Loader"

export interface IDataTableColumn {
  allowSort: boolean
  displayName: string
  render: (item: any) => React.ReactNode
  width?: number
}

export interface IProps {
  items: any[]
  columns: { [key: string]: IDataTableColumn }
  getItemKey: (item: any, index: number) => number
  onSortableHeaderClick: (sortKey: string) => void
  currentSortKey: string
  currentSortOrder: SortOrders
  currentPage: number
}

/**
 * The DataTable component takes a `columns` configuration prop and renders
 * the `items` prop into an HTML table.
 */
const DataTable: React.FunctionComponent<IProps> = ({
  columns,
  items,
  getItemKey,
  onSortableHeaderClick,
  currentSortKey,
  currentSortOrder,
}) => (
  <Table>
    <thead>
      <TableRow>
        {Object.keys(columns).map((key: string) => (
          <DataTableHeader
            key={key}
            isSortable={columns[key].allowSort}
            currentSortKey={currentSortKey}
            currentSortOrder={currentSortOrder}
            sortKey={key}
            width={columns[key].width || 1}
            onClick={() => {
              onSortableHeaderClick(key)
            }}
          >
            {columns[key].displayName}
          </DataTableHeader>
        ))}
      </TableRow>
    </thead>
    <tbody>
      {items &&
        items.map((item, i) => (
          <TableRow key={getItemKey(item, i)}>
            {Object.keys(columns).map(key => (
              <TableCell key={`${i}-${key}`} width={columns[key].width}>
                {columns[key].render(item)}
              </TableCell>
            ))}
          </TableRow>
        ))}
    </tbody>
  </Table>
)

export default DataTable
