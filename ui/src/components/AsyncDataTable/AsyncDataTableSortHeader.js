import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import { get } from "lodash"
import {
  TableHeaderCell,
  TableHeaderSortIcon,
  TableHeaderCellSortable,
} from "../styled/Table"

class AsyncDataTableSortHeader extends Component {
  constructor(props) {
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

    setQueryParams({
      sort_by: sortBy,
      order,
      page: 1,
    })
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
        direction={direction}
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

AsyncDataTableSortHeader.displayName = "AsyncDataTableSortHeader"

AsyncDataTableSortHeader.propTypes = {
  allowSort: PropTypes.bool.isRequired,
  displayName: PropTypes.node.isRequired,
  queryParams: PropTypes.object.isRequired,
  setQueryParams: PropTypes.func.isRequired,
  sortKey: PropTypes.string.isRequired,
  width: PropTypes.number.isRequired,
}

AsyncDataTableSortHeader.defaultProps = {
  allowSort: false,
  width: 1,
}

export default withQueryParams()(AsyncDataTableSortHeader)
