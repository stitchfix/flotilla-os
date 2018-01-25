import React, { Component } from "react"
import PropTypes from "prop-types"
import cn from "classnames"

export default class SortHeader extends Component {
  static propTypes = {
    currentSortKey: PropTypes.string,
    currentOrder: PropTypes.string,
    display: PropTypes.node,
    sortKey: PropTypes.string,
    updateQuery: PropTypes.func,
    style: PropTypes.object,
    className: PropTypes.string,
  }
  static defaultProps = {
    className: "",
  }
  constructor(props) {
    super(props)
    this.handleClick = this.handleClick.bind(this)
  }
  getNextSortState() {
    const { currentSortKey, currentOrder, sortKey } = this.props
    if (sortKey !== currentSortKey) {
      return {
        sort_by: sortKey,
        order: "asc",
      }
    } else if (sortKey === currentSortKey && currentOrder === "asc") {
      return {
        sort_by: sortKey,
        order: "desc",
      }
    }
    return {
      sort_by: null,
      order: null,
    }
  }
  handleClick() {
    const { sort_by, order } = this.getNextSortState()

    this.props.updateQuery([
      {
        key: "sort_by",
        value: sort_by,
        updateType: "SHALLOW",
      },
      {
        key: "order",
        value: order,
        updateType: "SHALLOW",
      },
      {
        key: "page",
        value: 1,
        updateType: "SHALLOW",
      },
    ])
  }
  render() {
    const { currentSortKey, currentOrder, sortKey, style } = this.props
    const className = cn({
      "pl-th": true,
      "pl-th-sort": true,
      "pl-th-sort-active": currentSortKey === sortKey,
      desc: currentSortKey === sortKey && currentOrder === "desc",
      asc: currentSortKey === sortKey && currentOrder === "asc",
    })

    return (
      <button
        onClick={this.handleClick}
        className={`${className} ${this.props.className}`}
        style={style}
      >
        {this.props.display}
      </button>
    )
  }
}
