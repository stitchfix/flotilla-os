import React, { Component } from "react"
import PropTypes from "prop-types"
import { Link } from "react-router-dom"
import withQueryParams from "react-router-query-params"
import { isEmpty } from "lodash"

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
    const { initialQuery, setQueryParams, queryParams } = this.props

    setQueryParams(initialQuery)
  }

  componentDidUpdate(prevProps) {
    // Check custom compare prop.
    if (this.props.shouldRequest(prevProps, this.props) === true) {
      this.requestData()
      return
    }

    // Check query parameters.
    this.compareQueryParams(prevProps.queryParams, this.props.queryParams)
  }

  getNecessaryQueryParamKeys() {
    return ["page", "order", "sort_by", ...Object.keys(this.props.filters)]
  }

  compareQueryParams(prev, curr) {
    // Compare necessary query params (page, filter, sort_by, order)
    const keys = this.getNecessaryQueryParamKeys()

    for (let i = 0; i < keys.length; i++) {
      if (prev[keys[i]] !== curr[keys[i]]) {
        this.requestData()
        break
      }
    }
  }

  requestData() {
    const { requestFn } = this.props

    requestFn(this.constructRequestQuery())
      .then(data => {
        this.setState({ data, requestState: requestStates.READY })
      })
      .catch(error => {
        this.setState({ error })
      })
  }

  constructRequestQuery() {
    const { queryParams, limit } = this.props
    const ks = this.getNecessaryQueryParamKeys()

    return Object.keys(queryParams).reduce((acc, key) => {
      if (key === "page") {
        acc.offset = AsyncDataTable.pageToOffset(queryParams[key], limit)
        acc.limit = limit
      } else if (ks.includes(key)) {
        acc[key] = queryParams[key]
      }
      return acc
    }, {})
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
  initialQuery: {
    page: 1,
  },
  limit: 20,
  requestFn: () => {},
  shouldRequest: (prevProps, currProps) => false,
}

export default withQueryParams()(AsyncDataTable)
