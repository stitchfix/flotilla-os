import React, { Component } from "react"
import PropTypes from "prop-types"

export default class PaginationButtons extends Component {
  static propTypes = {
    total: PropTypes.number,
    buttonCount: PropTypes.number,
    buttonEl: PropTypes.node,
    activeButtonClassName: PropTypes.string,
    offset: PropTypes.number,
    limit: PropTypes.number,
    updateQuery: PropTypes.func,
    wrapperEl: PropTypes.node,
  }
  static defaultProps = {
    buttonCount: 5,
    buttonEl: <button className="pl-button" />,
    activeButtonClassName: "active",
    wrapperEl: <div />,
  }
  offsetAndLimitToPage(offset, limit) {
    return +offset / +limit + 1
  }
  render() {
    const {
      total,
      buttonCount,
      buttonEl,
      activeButtonClassName,
      offset,
      limit,
      wrapperEl,
    } = this.props

    // Otherwise a `NaN` will be set as the child of the buttons and cause
    // an error to be logged.
    if (limit === undefined || offset === undefined) {
      return <span />
    }

    const currentPage = this.offsetAndLimitToPage(offset, limit)
    const totalPages = Math.ceil(total / limit)

    // Render less than props.buttonCount when totalPages < buttonCount
    const derivedButtonCount =
      buttonCount > totalPages ? totalPages : buttonCount

    return React.cloneElement(
      wrapperEl,
      {},
      [...Array(derivedButtonCount).keys()].map(n => {
        let pageNumber
        if (currentPage <= Math.ceil(buttonCount / 2)) {
          pageNumber = n + 1
        } else if (
          currentPage > Math.ceil(buttonCount / 2) &&
          currentPage + Math.floor(buttonCount / 2) <= totalPages
        ) {
          pageNumber = currentPage + n + 1 - Math.ceil(buttonCount / 2)
        } else {
          pageNumber = totalPages + n + 1 - buttonCount
        }

        // Append props.activeButtonClassName
        let className = buttonEl.props.className

        if (currentPage === pageNumber) {
          className += ` ${this.props.activeButtonClassName}`
        }

        return React.cloneElement(buttonEl, {
          children: pageNumber,
          key: n,
          className,
          onClick: () => {
            this.props.updateQuery({
              key: "page",
              value: pageNumber,
              updateType: "SHALLOW",
            })
          },
        })
      })
    )
  }
}
