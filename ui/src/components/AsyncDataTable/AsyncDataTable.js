import React, { Component } from "react"
import PropTypes from "prop-types"
import withQueryParams from "react-router-query-params"
import { get, isEmpty, omit, isObject, size, has, toString } from "lodash"

import Card from "../Card"
import EmptyTable from "../EmptyTable"

import AsyncDataTableFilter, {
  asyncDataTableFilterTypes,
} from "./AsyncDataTableFilter"
import AsyncDataTableSortHeader from "./AsyncDataTableSortHeader"
import AsyncDataTablePagination from "./AsyncDataTablePagination"

const requestStates = {
  READY: "READY",
  ERROR: "ERROR",
  NOT_READY: "NOT_READY",
}

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
    requestState: requestStates.NOT_READY,
    data: null,
    error: false,
    inFlight: false,
  }

  componentDidMount() {
    const { initialQuery, queryParams, setQueryParams } = this.props

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
      this.requestData(q)
    }
  }

  componentDidUpdate(prevProps) {
    const { queryParams } = this.props

    const prevQ = prevProps.queryParams
    const currQ = queryParams

    if (!this.areQueriesEqual(prevQ, currQ)) {
      this.requestData(currQ)
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
  requestData(query) {
    const { getRequestArgs, requestFn, limit } = this.props

    let q = omit(query, "page")
    q.offset = AsyncDataTable.pageToOffset(get(query, "page", 1), limit)
    q.limit = limit

    requestFn(getRequestArgs(q))
      .then(data => {
        this.setState({ data, requestState: requestStates.READY })
      })
      .catch(error => {
        this.setState({ error })
      })
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
    } = this.props
    const { requestState, data, error } = this.state

    switch (requestState) {
      case requestStates.ERROR:
        const errorDisplay = error.toString() || "An error occurred."
        return <EmptyTable title={errorDisplay} error />
      case requestStates.READY:
        const items = getItems(data)
        const total = getTotal(data)

        return (
          <div className="flot-list-view">
            {!isEmpty(filters) && (
              <Card
                className="flot-list-view-filters-container"
                contentStyle={{ padding: 0 }}
              >
                <div className="flot-list-view-filters">
                  {Object.keys(filters).map(key => (
                    <AsyncDataTableFilter
                      {...filters[key]}
                      filterKey={key}
                      key={key}
                    />
                  ))}
                </div>
              </Card>
            )}
            {isEmpty(items) ? (
              <EmptyTable title={emptyTableTitle} actions={emptyTableBody} />
            ) : (
              <div className="pl-table pl-bordered">
                <div className="pl-tr">
                  {Object.keys(columns).map(key => (
                    <AsyncDataTableSortHeader
                      allowSort={columns[key].allowSort}
                      displayName={columns[key].displayName}
                      sortKey={key}
                      key={key}
                      width={columns[key].width}
                    />
                  ))}
                </div>
                {items.map((item, i) => (
                  <div className="pl-tr hoverable" key={i}>
                    {Object.keys(columns).map(key => (
                      <div
                        className="pl-td"
                        key={`${i}-${key}`}
                        style={{ flex: get(columns[key], "width", 1) }}
                      >
                        {columns[key].render(item)}
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            )}
            <div
              className="table-footer"
              style={{
                display: "flex",
                flexFlow: "row nowrap",
                justifyContent: "center",
                alignItems: "center",
              }}
            >
              <AsyncDataTablePagination total={total} limit={limit} />
            </div>
          </div>
        )
      case requestStates.NOT_READY:
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
  getItems: PropTypes.func.isRequired,
  getRequestArgs: PropTypes.func.isRequired,
  getTotal: PropTypes.func.isRequired,
  initialQuery: PropTypes.object,
  limit: PropTypes.number.isRequired,
  queryParams: PropTypes.object.isRequired,
  requestFn: PropTypes.func.isRequired,
  setQueryParams: PropTypes.func.isRequired,
  shouldRequest: PropTypes.func.isRequired,
}

AsyncDataTable.defaultProps = {
  children: () => <span />,
  columns: {},
  emptyTableBody: "",
  emptyTableTitle: "This collection is empty",
  filters: {},
  getItems: data => [],
  getRequestArgs: query => query,
  initialQuery: {},
  limit: 20,
  requestFn: () => {},
  shouldRequest: (prevProps, currProps) => false,
}

export default withQueryParams()(AsyncDataTable)
