import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
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

    if (!this.areQueriesEqual(initialQuery, queryParams)) {
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
    const { requestFn, limit } = this.props

    let q = omit(query, "page")
    q.offset = AsyncDataTable.pageToOffset(get(query, "page", 1), limit)
    q.limit = limit

    requestFn(q)
      .then(data => {
        this.setState({ data, requestState: requestStates.READY })
      })
      .catch(error => {
        this.setState({ error })
      })
  }

  render() {
    const { columns, filters, getItems, getTotal, limit } = this.props
    const { requestState, data } = this.state

    switch (requestState) {
      case requestStates.ERROR:
        return "uh oh"
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
            <div className="pl-table pl-bordered">
              <div className="pl-tr">
                {Object.keys(columns).map(key => {
                  const col = columns[key]

                  if (col.allowSort) {
                    return (
                      <AsyncDataTableSortHeader
                        displayName={col.displayName}
                        sortKey={key}
                        key={key}
                      />
                    )
                  }

                  return (
                    <div className="pl-th" key={key}>
                      {col.displayName}
                    </div>
                  )
                })}
              </div>
              {items.map((item, i) => (
                <div className="pl-tr" key={i}>
                  {Object.keys(columns).map(key => (
                    <div className="pl-td" key={`${i}-${key}`}>
                      {columns[key].render(item)}
                    </div>
                  ))}
                </div>
              ))}
            </div>
            <div
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
        return (
          <EmptyTable
            title="No tasks were found. Create one?"
            actions={
              <Link className="pl-button pl-intent-primary" to="/tasks/create">
                <span style={{ marginLeft: 4 }}>Create New Task</span>
              </Link>
            }
          />
        )
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
    })
  ).isRequired,
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
  filters: {},
  getItems: data => [],
  initialQuery: {},
  limit: 20,
  requestFn: () => {},
  shouldRequest: (prevProps, currProps) => false,
}

export default withQueryParams()(AsyncDataTable)