import * as React from "react"
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
import AsyncDataTableFilter from "./AsyncDataTableFilter"
import AsyncDataTableSortHeader from "./AsyncDataTableSortHeader"
import AsyncDataTablePagination from "./AsyncDataTablePagination"
import {
  AsyncDataTableContainer,
  AsyncDataTableFilters,
  AsyncDataTableContent,
  AsyncDataTableLoadingMask,
} from "../styled/AsyncDataTable"
import config from "../../config"
import PopupContext from "../Popup/PopupContext"
import QueryParams from "../QueryParams/QueryParams"
import Loader from "../styled/Loader"
import {
  IFlotillaUIAsyncDataTableFilterProps,
  flotillaUIRequestStates,
  IFlotillaUIPopupProps,
  flotillaUIIntents,
  IFlotillaAPIError,
} from "../../.."

interface IAsyncDataTableColumn {
  allowSort: boolean
  displayName: string
  render: (item: any) => React.ReactNode
  width?: number
}

interface IUnwrappedAsyncDataTableProps {
  columns: { [key: string]: IAsyncDataTableColumn }
  emptyTableBody?: React.ReactNode
  emptyTableTitle: string
  filters?: { [key: string]: IFlotillaUIAsyncDataTableFilterProps }
  getItemKey: (item: any, index: number) => number
  getItems: (data: any) => any[]
  getRequestArgs: (query: any) => any
  getTotal: (data: any) => number
  initialQuery: object
  isView: boolean
  limit: number
  requestFn: (arg: any) => any
  shouldContinuouslyFetch: boolean
}

interface IAsyncDataTableProps extends IUnwrappedAsyncDataTableProps {
  renderPopup: (p: IFlotillaUIPopupProps) => void
  queryParams: any
  setQueryParams: (query: object, shouldReplace: boolean) => void
}

interface IAsyncDataTableState {
  requestState: flotillaUIRequestStates
  data: any[]
  error: any
  inFlight: boolean
}

/**
 * AsyncDataTable takes a requestFn prop (usually a bound method of the
 * FlotillaAPIClient, e.g. getTasks) and requests and renders the data.
 * Additionally, it will handle pagination, filters, and sorting via the router
 * query.
 */
class AsyncDataTable extends React.PureComponent<
  IAsyncDataTableProps,
  IAsyncDataTableState
> {
  static displayName = "AsyncDataTable"
  static defaultProps: Partial<IAsyncDataTableProps> = {
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
  }
  static offsetToPage = (offset: any, limit: any): number =>
    +offset / +limit + 1
  static pageToOffset = (page: any, limit: any): number => (+page - 1) * +limit

  private requestInterval: number | null = null

  state = {
    requestState: flotillaUIRequestStates.NOT_READY,
    data: [],
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
        if (document.visibilityState !== "visible") return

        this.requestData()
      }, +config.RUN_REQUEST_INTERVAL_MS)
    }
  }

  componentDidUpdate(prevProps: IAsyncDataTableProps) {
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

  /** Performs a shallow comparison of two query objects. */
  areQueriesEqual(
    a: { [key: string]: any },
    b: { [key: string]: any }
  ): boolean {
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

  /** Sends a request via the requestFn prop then stores data in state. */
  requestData(): void {
    if (this.state.inFlight === true) return

    this.setState({ inFlight: true, error: false })

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
      .then((data: any) => {
        this.setState({
          data,
          requestState: flotillaUIRequestStates.READY,
          inFlight: false,
        })
      })
      .catch((error: IFlotillaAPIError) => {
        this.clearInterval()

        this.props.renderPopup({
          body: error.data,
          intent: flotillaUIIntents.ERROR,
          shouldAutohide: false,
          title: `An error occurred (Status Code: ${error.status})`,
        })

        this.setState({ error, inFlight: false })
      })
  }

  clearInterval = () => {
    if (this.requestInterval !== null)
      window.clearInterval(this.requestInterval)
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

    const { requestState, data, error, inFlight } = this.state

    switch (requestState) {
      case flotillaUIRequestStates.ERROR:
        const errorDisplay = error.toString() || "An error occurred."
        return <EmptyTable title={errorDisplay} error />
      case flotillaUIRequestStates.READY:
        const items = getItems(data)
        const total = getTotal(data)

        return (
          <AsyncDataTableContainer>
            {!!inFlight && (
              <AsyncDataTableLoadingMask>
                <Loader intent={flotillaUIIntents.PRIMARY} />
              </AsyncDataTableLoadingMask>
            )}
            {!!filters &&
              !isEmpty(filters) && (
                <AsyncDataTableFilters isView={isView}>
                  {Object.keys(filters).map(key => (
                    <AsyncDataTableFilter
                      {...filters[key]}
                      name={key}
                      key={key}
                    />
                  ))}
                </AsyncDataTableFilters>
              )}
            <AsyncDataTableContent>
              {isEmpty(items) ? (
                <EmptyTable title={emptyTableTitle} actions={emptyTableBody} />
              ) : (
                <React.Fragment>
                  <Table>
                    <TableRow>
                      {Object.keys(columns).map(key => (
                        <AsyncDataTableSortHeader
                          allowSort={columns[key].allowSort}
                          displayName={columns[key].displayName}
                          sortKey={key}
                          key={key}
                          width={columns[key].width || 1}
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
                  <AsyncDataTablePagination total={total} limit={limit} />
                </React.Fragment>
              )}
            </AsyncDataTableContent>
          </AsyncDataTableContainer>
        )
      case flotillaUIRequestStates.NOT_READY:
      default:
        return <EmptyTable isLoading />
    }
  }
}

const WrappedAsyncDataTable: React.SFC<
  IUnwrappedAsyncDataTableProps
> = props => (
  <PopupContext.Consumer>
    {ctx => (
      <QueryParams>
        {({ queryParams, setQueryParams }) => (
          <AsyncDataTable
            {...props}
            renderPopup={ctx.renderPopup}
            queryParams={queryParams}
            setQueryParams={setQueryParams}
          />
        )}
      </QueryParams>
    )}
  </PopupContext.Consumer>
)

export default WrappedAsyncDataTable
