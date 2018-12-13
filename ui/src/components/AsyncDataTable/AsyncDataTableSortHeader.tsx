import * as React from "react"
import { get } from "lodash"
import {
  TableHeaderCell,
  TableHeaderSortIcon,
  TableHeaderCellSortable,
} from "../styled/Table"
import QueryParams from "../QueryParams/QueryParams"

interface IUnwrappedAsyncDataTableSortHeaderProps {
  allowSort: boolean
  displayName: React.ReactNode
  sortKey: string
  width: number
}

interface IAsyncDataTableSortHeaderProps
  extends IUnwrappedAsyncDataTableSortHeaderProps {
  queryParams: any
  setQueryParams: (query: object, shouldReplace: boolean) => void
}

class AsyncDataTableSortHeader extends React.PureComponent<
  IAsyncDataTableSortHeaderProps
> {
  static displayName = "AsyncDataTableSortHeader"

  static defaultProps: Partial<IAsyncDataTableSortHeaderProps> = {
    allowSort: false,
    width: 1,
  }

  constructor(props: IAsyncDataTableSortHeaderProps) {
    super(props)
    this.getCurrSortKey = this.getCurrSortKey.bind(this)
    this.getCurrSortOrder = this.getCurrSortOrder.bind(this)
    this.getNextSortState = this.getNextSortState.bind(this)
    this.handleClick = this.handleClick.bind(this)
  }

  getCurrSortKey() {
    return get(this.props.queryParams, "sort_by", null)
  }

  getCurrSortOrder() {
    return get(this.props.queryParams, "order", null)
  }

  getNextSortState() {
    const { sortKey } = this.props
    const currSortKey = this.getCurrSortKey()
    const currSortOrder = this.getCurrSortOrder()

    if (sortKey !== currSortKey) {
      return {
        sortBy: sortKey,
        order: "asc",
      }
    }

    if (sortKey === currSortKey && currSortOrder === "asc") {
      return {
        sortBy: sortKey,
        order: "desc",
      }
    }

    return {
      sortBy: null,
      order: null,
    }
  }

  handleClick() {
    const { setQueryParams } = this.props
    const { sortBy, order } = this.getNextSortState()

    setQueryParams(
      {
        sort_by: sortBy,
        order,
        page: 1,
      },
      false
    )
  }

  render() {
    const { allowSort, displayName, sortKey, width } = this.props

    if (allowSort !== true) {
      return <TableHeaderCell width={width}>{displayName}</TableHeaderCell>
    }
    const currSortKey = this.getCurrSortKey()
    const currSortOrder = this.getCurrSortOrder()

    const isActive = currSortKey === sortKey
    let direction = null

    if (isActive) {
      direction = currSortOrder
    }

    return (
      <TableHeaderCellSortable
        onClick={this.handleClick}
        width={width}
        isActive={isActive}
      >
        {displayName}
        {!!isActive &&
          !!direction && (
            <TableHeaderSortIcon>
              {direction === "asc" ? "▲" : "▼"}
            </TableHeaderSortIcon>
          )}
      </TableHeaderCellSortable>
    )
  }
}

const WrappedAsyncDataTableSortHeader: React.SFC<
  IUnwrappedAsyncDataTableSortHeaderProps
> = props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => (
      <AsyncDataTableSortHeader
        {...props}
        queryParams={queryParams}
        setQueryParams={setQueryParams}
      />
    )}
  </QueryParams>
)

export default WrappedAsyncDataTableSortHeader
