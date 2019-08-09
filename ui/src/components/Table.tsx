import * as React from "react"
import { HTMLTable } from "@blueprintjs/core"
import { isEmpty, isArray } from "lodash"
import SortableTh from "./SortableTh"
import { SortOrder } from "../types"

type Column = {
  displayName: string
  render: (item: any) => React.ReactNode
  isSortable: boolean
}

type Props<T> = {
  items: T[]
  columns: { [key: string]: Column }
  getItemKey: (item: T, index: number) => any
  updateSort: (sortKey: string) => void
  currentSortKey: string
  currentSortOrder: SortOrder
}

class Table<T> extends React.Component<Props<T>> {
  render() {
    const {
      columns,
      items,
      getItemKey,
      updateSort,
      currentSortKey,
      currentSortOrder,
    } = this.props

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
          {isArray(items) &&
            !isEmpty(items) &&
            items.map((item, i) => (
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
}

export default Table
