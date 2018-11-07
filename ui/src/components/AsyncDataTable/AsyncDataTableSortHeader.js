import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import { get } from "lodash"
import cn from "classnames"

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
    const { displayName, sortKey } = this.props
    const currSortKey = this.getCurrSortKey()
    const currSortOrder = this.getCurrSortOrder()

    const className = cn({
      "pl-th": true,
      "pl-th-sort": true,
      "pl-th-sort-active": currSortKey === sortKey,
      desc: currSortKey === sortKey && currSortOrder === "desc",
      asc: currSortKey === sortKey && currSortOrder === "asc",
    })

    return (
      <button onClick={this.handleClick} className={className}>
        {displayName}
      </button>
    )
  }
}

AsyncDataTableSortHeader.propTypes = {
  displayName: PropTypes.node.isRequired,
  queryParams: PropTypes.object.isRequired,
  setQueryParams: PropTypes.func.isRequired,
  sortKey: PropTypes.string.isRequired,
}
AsyncDataTableSortHeader.defaultProps = {}

export default withQueryParams()(AsyncDataTableSortHeader)
