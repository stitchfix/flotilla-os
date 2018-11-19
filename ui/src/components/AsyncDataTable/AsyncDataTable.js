import React, { Component, createRef } from "react"
import PropTypes from "prop-types"
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
import PopupContext from "../Popup/PopupContext"
import intentTypes from "../../constants/intentTypes"
import QueryParams from "../QueryParams/QueryParams"
import {
  transformReactFormValuesToQueryParams,
  transformQueryParamsToReactFormValues,
} from "../../utils/reactFormQueryParams"

const FORM_REF = createRef()

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
      setQueryParams(q, true)
    } else {
      this.requestData()
    }

    if (shouldContinuouslyFetch) {
      this.requestInterval = window.setInterval(() => {
        // Return if the browser tab isn't focused.
        if (!this.props.isTabFocused) return

        this.requestData()
      }, config.RUN_REQUEST_INTERVAL_MS)
    }
  }

  componentDidUpdate(prevProps) {
    const { queryParams } = this.props

    const prevQ = prevProps.queryParams
    const currQ = queryParams

    if (!this.areQueriesEqual(prevQ, currQ)) {
      // @TODO: this is a hack to sync React Form values (i.e. the filters)
      // with the query string. Moving forward, we should remove React Form
      // from this component and only rely and query params as the source of
      // truth.
      if (has(this.state, ["formAPI", "setAllValues"])) {
        this.state.formAPI.setAllValues(
          transformQueryParamsToReactFormValues(currQ)
        )
      }
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
    // Return false if the arguments are not objects.
    if (!isObject(a) || !isObject(b)) {
      return false
    }

    // Return false if the size differs.
    if (size(a) !== size(b)) {
      return false
    }

    // Perform shallow comparison.
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
    const {
      queryParams,
      getRequestArgs,
      requestFn,
      limit,
      renderPopup,
    } = this.props

    let q = omit(queryParams, "page")
    q.offset = AsyncDataTable.pageToOffset(get(queryParams, "page", 1), limit)
    q.limit = limit

    requestFn(getRequestArgs(q))
      .then(data => {
        this.setState({ data, requestState: requestStateTypes.READY })
      })
      .catch(error => {
        this.setState({ error })
        renderPopup({
          title: "An error occurred",
          body: get(error, ["response", "data"], toString(error)),
          intent: intentTypes.error,
        })
      })
  }

  handleFiltersChange = (formState, formAPI) => {
    const { setQueryParams } = this.props
    const q = transformReactFormValuesToQueryParams(
      get(formState, "values", {})
    )

    setQueryParams(q)
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
      isView,
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
                defaultValues={transformQueryParamsToReactFormValues(
                  queryParams
                )}
                getApi={formAPI => {
                  this.setState({ formAPI })
                }}
              >
                {formAPI => {
                  return (
                    <AsyncDataTableFilters isView={isView}>
                      {Object.keys(filters).map(key => (
                        <AsyncDataTableFilter
                          {...filters[key]}
                          formAPI={formAPI}
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
  isTabFocused: PropTypes.bool.isRequired,
  isView: PropTypes.bool.isRequired,
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
  isView: true,
  limit: 50,
  requestFn: () => {},
  shouldContinuouslyFetch: false,
  shouldRequest: (prevProps, currProps) => false,
}

export default props => (
  <PageVisibility>
    {isTabFocused => (
      <PopupContext.Consumer>
        {ctx => (
          <QueryParams>
            {({ queryParams, setQueryParams }) => (
              <AsyncDataTable
                {...props}
                isTabFocused={isTabFocused}
                renderPopup={ctx.renderPopup}
                queryParams={queryParams}
                setQueryParams={setQueryParams}
              />
            )}
          </QueryParams>
        )}
      </PopupContext.Consumer>
    )}
  </PageVisibility>
)
