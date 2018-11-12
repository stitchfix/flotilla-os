import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import { get } from "lodash"
import Button from "../styled/Button"
import ButtonGroup from "../styled/ButtonGroup"

class AsyncDataTablePagination extends Component {
  constructor(props) {
    super(props)
    this.handlePrevClick = this.handlePrevClick.bind(this)
    this.handleNextClick = this.handleNextClick.bind(this)
    this.handleFirstClick = this.handleFirstClick.bind(this)
    this.handleLastClick = this.handleLastClick.bind(this)
    this.getCurrPage = this.getCurrPage.bind(this)
    this.updateQuery = this.updateQuery.bind(this)
    this.isFirstPage = this.isFirstPage.bind(this)
    this.isLastPage = this.isLastPage.bind(this)
  }

  handlePrevClick() {
    this.updateQuery(this.getCurrPage() - 1)
  }

  handleNextClick() {
    this.updateQuery(this.getCurrPage() + 1)
  }

  handleFirstClick() {
    this.updateQuery(1)
  }

  handleLastClick() {
    const { total, limit } = this.props
    this.updateQuery((total - total % limit) / limit)
  }

  getCurrPage() {
    return +get(this.props.queryParams, "page", 1)
  }

  updateQuery(page) {
    this.props.setQueryParams({ page })
  }

  isFirstPage() {
    return this.getCurrPage() === 1
  }

  isLastPage() {
    const { limit, total } = this.props
    return this.getCurrPage() * limit + limit > total
  }

  render() {
    const isFirstPage = this.isFirstPage()
    const isLastPage = this.isLastPage()
    return (
      <ButtonGroup>
        <Button
          key="first"
          disabled={isFirstPage}
          onClick={this.handleFirstClick}
        >
          First
        </Button>
        <Button
          key="prev"
          disabled={isFirstPage}
          onClick={this.handlePrevClick}
        >
          Prev
        </Button>
        <Button key="next" disabled={isLastPage} onClick={this.handleNextClick}>
          Next
        </Button>
        <Button key="last" disabled={isLastPage} onClick={this.handleLastClick}>
          Last
        </Button>
      </ButtonGroup>
    )
  }
}

AsyncDataTablePagination.displayName = "AsyncDataTablePagination"

AsyncDataTablePagination.propTypes = {
  limit: PropTypes.number.isRequired,
  queryParams: PropTypes.object.isRequired,
  setQueryParams: PropTypes.func.isRequired,
  total: PropTypes.number,
}

AsyncDataTablePagination.defaultProps = {}

export default withQueryParams()(AsyncDataTablePagination)
