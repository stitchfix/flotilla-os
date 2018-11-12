import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import { Form as ReactForm } from "react-form"
import PageVisibility from "react-page-visibility"
import {
  get,
  isEmpty,
  omit,
  isObject,
  size,
  has,
  toString,
  isNumber,
} from "lodash"
import EmptyTable from "../styled/EmptyTable"
import { Table, TableRow, TableCell } from "../styled/Table"
import AsyncDataTableFilter, {
  asyncDataTableFilterTypes,
} from "./AsyncDataTableFilter"
import AsyncDataTableSortHeader from "./AsyncDataTableSortHeader"
import AsyncDataTablePagination from "./AsyncDataTablePagination"
import * as requestStateTypes from "../../constants/requestStateTypes"
import {
  AsyncDataTableContainer,
  AsyncDataTableFilters,
  AsyncDataTableContent,
} from "../styled/AsyncDataTable"
import config from "../../config"

/**
 * AsyncDataTable takes a requestFn prop (usually a bound method of the
 * FlotillaAPIClient, e.g. getTasks) and requests and renders the data.
 * Additionally, it will handle pagination, filters, and sorting via the router
 * query.
 */
class AsyncDataTable extends Component {
  static offsetToPage = (offset, limit) => +offset / +limit + 1

  static pageToOffset = (page, limit) => (+page - 1) * +limit

  state = {
    requestState: requestStateTypes.NOT_READY,
    data: null,
    error: false,
    inFlight: false,
  }

  componentDidMount() {
    const {
      initialQuery,
      queryParams,
      setQueryParams,
      shouldContinuouslyFetch,
    } = this.props

    const q = {
      ...initialQuery,
      ...queryParams,
    }

    if (
      isEmpty(queryParams) &&
      !this.areQueriesEqual(initialQuery, queryParams)
    ) {
      setQueryParams(q)
    } else {
      this.requestData()
    }

    if (shouldContinuouslyFetch) {
      this.requestInterval = window.setInterval(() => {
        this.requestData()
      }, config.RUN_REQUEST_INTERVAL_MS)
    }
  }

  componentDidUpdate(prevProps) {
    const { queryParams } = this.props

    const prevQ = prevProps.queryParams
    const currQ = queryParams

    if (!this.areQueriesEqual(prevQ, currQ)) {
      this.requestData()
    }
  }

  componentWillUnmount() {
    if (isNumber(this.requestInterval)) {
      window.clearInterval(this.requestInterval)
    }
  }

  /**
   * Performs a shallow comparison of two query objects.
   *
   * @param {object} a
   * @param {object} b
   * @returns {boolean}
   */
  areQueriesEqual(a = {}, b = {}) {
    if (!isObject(a) || !isObject(b)) {
      return false
    }

    if (size(a) !== size(b)) {
      return false
    }

    for (let key in a) {
      if (!has(b, key)) {
        return false
      }

      if (toString(a[key]) !== toString(b[key])) {
        return false
      }
    }

    return true
  }

  /**
   * Sends a request via the requestFn prop then stores data in state.
   *
   * @param {object} query
   */
  requestData() {
    const { queryParams, getRequestArgs, requestFn, limit } = this.props

    let q = omit(queryParams, "page")
    q.offset = AsyncDataTable.pageToOffset(get(queryParams, "page", 1), limit)
    q.limit = limit

    requestFn(getRequestArgs(q))
      .then(data => {
        this.setState({ data, requestState: requestStateTypes.READY })
      })
      .catch(error => {
        this.setState({ error })
      })
  }

  handleFiltersChange = (formState, formAPI) => {
    const { setQueryParams } = this.props

    setQueryParams(get(formState, "values", {}))
  }

  render() {
    const {
      columns,
      filters,
      getItems,
      getTotal,
      limit,
      emptyTableBody,
      emptyTableTitle,
      queryParams,
      getItemKey,
    } = this.props
    const { requestState, data, error } = this.state

    switch (requestState) {
      case requestStateTypes.ERROR:
        const errorDisplay = error.toString() || "An error occurred."
        return <EmptyTable title={errorDisplay} error />
      case requestStateTypes.READY:
        const items = getItems(data)
        const total = getTotal(data)

        return (
          <AsyncDataTableContainer>
            {!isEmpty(filters) && (
              <ReactForm
                onChange={this.handleFiltersChange}
                defaultValues={queryParams}
              >
                {formAPI => {
                  return (
                    <AsyncDataTableFilters>
                      {Object.keys(filters).map(key => (
                        <AsyncDataTableFilter
                          {...filters[key]}
                          field={key}
                          key={key}
                        />
                      ))}
                    </AsyncDataTableFilters>
                  )
                }}
              </ReactForm>
            )}
            <AsyncDataTableContent>
              {isEmpty(items) ? (
                <EmptyTable title={emptyTableTitle} actions={emptyTableBody} />
              ) : (
                <Table>
                  <TableRow>
                    {Object.keys(columns).map(key => (
                      <AsyncDataTableSortHeader
                        allowSort={columns[key].allowSort}
                        displayName={columns[key].displayName}
                        sortKey={key}
                        key={key}
                        width={columns[key].width}
                      />
                    ))}
                  </TableRow>
                  {items.map((item, i) => (
                    <TableRow key={getItemKey(item, i)}>
                      {Object.keys(columns).map(key => (
                        <TableCell
                          key={`${i}-${key}`}
                          width={columns[key].width}
                        >
                          {columns[key].render(item)}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))}
                </Table>
              )}
              <AsyncDataTablePagination total={total} limit={limit} />
            </AsyncDataTableContent>
          </AsyncDataTableContainer>
        )
      case requestStateTypes.NOT_READY:
      default:
        return <EmptyTable isLoading />
    }
  }
}

AsyncDataTable.displayName = "AsyncDataTable"

AsyncDataTable.propTypes = {
  children: PropTypes.func.isRequired,
  columns: PropTypes.objectOf(
    PropTypes.shape({
      allowSort: PropTypes.bool.isRequired,
      displayName: PropTypes.string.isRequired,
      render: PropTypes.func.isRequired,
      width: PropTypes.number,
    })
  ).isRequired,
  emptyTableBody: PropTypes.node,
  emptyTableTitle: PropTypes.string.isRequired,
  filters: PropTypes.objectOf(
    PropTypes.shape({
      displayName: PropTypes.string.isRequired,
      type: PropTypes.oneOf(Object.values(asyncDataTableFilterTypes))
        .isRequired,
      options: PropTypes.arrayOf(
        PropTypes.shape({
          label: PropTypes.string,
          value: PropTypes.string,
        })
      ),
    })
  ),
  getItemKey: PropTypes.func.isRequired,
  getItems: PropTypes.func.isRequired,
  getRequestArgs: PropTypes.func.isRequired,
  getTotal: PropTypes.func.isRequired,
  initialQuery: PropTypes.object,
  limit: PropTypes.number.isRequired,
  queryParams: PropTypes.object.isRequired,
  requestFn: PropTypes.func.isRequired,
  setQueryParams: PropTypes.func.isRequired,
  shouldContinuouslyFetch: PropTypes.bool.isRequired,
  shouldRequest: PropTypes.func.isRequired,
}

AsyncDataTable.defaultProps = {
  children: () => <span />,
  columns: {},
  emptyTableBody: "",
  emptyTableTitle: "This collection is empty",
  filters: {},
  getItemKey: (item, index) => index,
  getItems: data => [],
  getRequestArgs: query => query,
  initialQuery: {},
  limit: 50,
  requestFn: () => {},
  shouldContinuouslyFetch: false,
  shouldRequest: (prevProps, currProps) => false,
}

export default withQueryParams()(props => (
  <PageVisibility>
    {isVisible => <AsyncDataTable {...props} isVisible={isVisible} />}
  </PageVisibility>
))
