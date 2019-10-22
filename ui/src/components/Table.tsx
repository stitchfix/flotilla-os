import * as React from "react"
import { HTMLTable, Callout } from "@blueprintjs/core"
import { isArray } from "lodash"
import SortableTh from "./SortableTh"
import { SortOrder } from "../types"

type Column<ItemType> = {
  displayName: string
  render: (item: ItemType) => React.ReactNode
  isSortable: boolean
}

type Props<ItemType> = {
  items: ItemType[]
  columns: { [key: string]: Column<ItemType> }
  getItemKey: (item: ItemType, index: number) => any
  updateSort: (sortKey: string) => void
  currentSortKey: string
  currentSortOrder: SortOrder
}

class Table<ItemType> extends React.Component<Props<ItemType>> {
  render() {
    const {
      columns,
      items,
      getItemKey,
      updateSort,
      currentSortKey,
      currentSortOrder,
    } = this.props

    if (isArray(items) && items.length > 0) {
      return (
        <HTMLTable striped bordered>
          <thead>
            <tr>
              {Object.entries(columns).map(([k, v]) => (
                <SortableTh
                  isSortable={v.isSortable}
                  isActive={currentSortKey === k}
                  order={currentSortOrder}
                  onClick={updateSort.bind(this, k)}
                  key={k}
                >
                  {v.displayName}
                </SortableTh>
              ))}
            </tr>
          </thead>
          <tbody>
            {items.map((item, i) => (
              <tr key={getItemKey(item, i)}>
                {Object.entries(columns).map(([k, v]) => (
                  <td key={k}>{v.render(item)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </HTMLTable>
      )
    }

    return <Callout>No items were found.</Callout>
  }
}

export default Table
