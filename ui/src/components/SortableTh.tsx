import * as React from "react"
import { SortOrder } from "../types"

export type Props = {
  isSortable: boolean
  isActive: boolean
  order: SortOrder
  onClick: () => void
}

const Th: React.FunctionComponent<Props> = ({
  isSortable,
  isActive,
  order,
  children,
  onClick,
}) => {
  let className = ""

  if (isSortable) {
    className += "flotilla-th-sortable"
    if (isActive) {
      className += " active"

      if (order === SortOrder.ASC) {
        className += " active-asc"
      } else {
        className += " active-desc"
      }
    }
  }

  return (
    <th onClick={onClick} className={className}>
      {children}
    </th>
  )
}

export default Th
