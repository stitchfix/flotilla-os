import * as React from "react"
import styled from "styled-components"
import { get } from "lodash"
import Button from "../styled/Button"
import { SPACING_PX } from "../../helpers/styles"
import QueryParams from "../QueryParams/QueryParams"

const PaginationButtonGroup = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: center;
  align-items: center;
  padding: ${SPACING_PX * 2}px 0;
  width: 100%;

  & > * {
    margin: 0 ${SPACING_PX}px;
  }
`

interface IUnwrappedAsyncDataTablePaginationProps {
  limit: number
  total: number
}

interface IAsyncDataTablePaginationProps
  extends IUnwrappedAsyncDataTablePaginationProps {
  queryParams: any
  setQueryParams: (query: object, shouldReplace: boolean) => void
}

class AsyncDataTablePagination extends React.PureComponent<
  IAsyncDataTablePaginationProps
> {
  static displayName = "AsyncDataTablePagination"
  static defaultProps: Partial<IAsyncDataTablePaginationProps> = {
    total: 0,
  }
  handlePrevClick = (): void => {
    this.updateQuery(this.getCurrPage() - 1)
  }

  handleNextClick = (): void => {
    this.updateQuery(this.getCurrPage() + 1)
  }

  handleFirstClick = (): void => {
    this.updateQuery(1)
  }

  handleLastClick = (): void => {
    const { total, limit } = this.props
    this.updateQuery((total - total % limit) / limit)
  }

  getCurrPage = (): number => {
    return +get(this.props.queryParams, "page", 1)
  }

  updateQuery = (page: number): void => {
    this.props.setQueryParams({ page }, false)
  }

  isFirstPage = (): boolean => {
    return this.getCurrPage() === 1
  }

  isLastPage = (): boolean => {
    const { limit, total } = this.props
    return this.getCurrPage() * limit + limit > total
  }

  render() {
    const isFirstPage = this.isFirstPage()
    const isLastPage = this.isLastPage()
    return (
      <PaginationButtonGroup>
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
      </PaginationButtonGroup>
    )
  }
}

const WrappedAsyncDataTablePagination: React.SFC<
  IUnwrappedAsyncDataTablePaginationProps
> = props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => (
      <AsyncDataTablePagination
        {...props}
        queryParams={queryParams}
        setQueryParams={setQueryParams}
      />
    )}
  </QueryParams>
)

export default WrappedAsyncDataTablePagination
