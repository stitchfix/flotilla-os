import * as React from "react"
import { get, isEmpty, omit, isNumber } from "lodash"
import config from "../../config"
import QueryParams from "../QueryParams/QueryParams"
import Request, { IChildProps as IRequestChildProps } from "../Request/Request"
import areObjectsEqualShallow from "../../helpers/areObjectsEqualShallow"
import { flotillaUIRequestStates } from "../../types"

export enum SortOrders {
  ASC = "asc",
  DESC = "desc",
}

// Prop shape for the connected (exported) component.
interface IConnectedProps {
  getRequestArgs: (query: any) => any
  initialQuery: object
  limit: number
  requestFn: (arg: any) => any
  shouldContinuouslyFetch: boolean
  children: (props: IChildProps) => React.ReactNode
}

// Prop shape for the unconnected component.
export interface IProps extends IConnectedProps {
  data: any
  inFlight: boolean
  requestState: flotillaUIRequestStates
  request: (args?: any) => void
  error: any
  queryParams: any
  setQueryParams: (query: object, shouldReplace: boolean) => void
}

export interface IChildProps {
  queryParams: any
  requestState: flotillaUIRequestStates
  error: any
  inFlight: boolean
  data: any
  updatePage: (page: number) => void
  updateSort: (key: string) => void
  updateSearch: (key: string, value: any) => void
  currentSortKey: string
  currentSortOrder: SortOrders
  currentPage: number
}

/**
 * ListRequest takes a requestFn prop (usually a bound method of the
 * FlotillaAPIClient, e.g. getTasks) and requests and renders the data.
 * Additionally, it will handle pagination, filters, and sorting via the URL
 * query.
 */
export class ListRequest extends React.PureComponent<IProps> {
  static displayName = "ListRequest"
  static defaultProps: Partial<IProps> = {
    getRequestArgs: query => query,
    initialQuery: {},
    limit: 50,
    requestFn: () => {},
    shouldContinuouslyFetch: false,
  }
  static offsetToPage = (offset: any, limit: any): number =>
    +offset / +limit + 1
  static pageToOffset = (page: any, limit: any): number => (+page - 1) * +limit
  static preprocessQuery(
    query: { [key: string]: any },
    limit: number
  ): { [key: string]: any } {
    let q = omit(query, "page")
    q.offset = ListRequest.pageToOffset(get(query, "page", 1), limit)
    q.limit = limit
    return q
  }
  requestInterval: number | null = null

  componentDidMount() {
    const {
      initialQuery,
      queryParams,
      setQueryParams,
      shouldContinuouslyFetch,
    } = this.props

    if (isEmpty(queryParams)) {
      setQueryParams(initialQuery, true)
    } else {
      this.requestData()
    }

    if (shouldContinuouslyFetch) {
      this.requestInterval = window.setInterval(() => {
        // Return if the browser tab isn't focused to prevent unnecessary calls.
        if (document.visibilityState !== "visible") return

        this.requestData()
      }, +config.RUN_REQUEST_INTERVAL_MS)
    }
  }

  componentDidUpdate(prevProps: IProps) {
    if (
      !areObjectsEqualShallow(prevProps.queryParams, this.props.queryParams)
    ) {
      this.requestData()
    }
  }

  componentWillUnmount() {
    this.clearInterval()
  }

  /** Sends a request via the requestFn prop then stores data in state. */
  requestData(): void {
    const { request, getRequestArgs, queryParams, limit } = this.props
    request(getRequestArgs(ListRequest.preprocessQuery(queryParams, limit)))
  }

  /** Clears this.requestInterval if set. */
  clearInterval(): void {
    if (isNumber(this.requestInterval))
      window.clearInterval(this.requestInterval)
  }

  /** Updates the query's `page` attribute. */
  updatePage(page: number): void {
    if (page >= 1) {
      this.props.setQueryParams({ page }, false)
    }
  }

  /** Updates any attribute in the query. */
  updateSearch(key: string, value: string | number | boolean): void {
    this.props.setQueryParams({ [key]: value }, false)
  }

  /** Updates the query's `sort_by` and `order` attributes. */
  updateSort(sortKey: string): void {
    const { queryParams, setQueryParams } = this.props
    let nextSort: { sort_by: string; order: SortOrders } = {
      sort_by: sortKey,
      order: SortOrders.ASC,
    }

    const currSortKey = get(queryParams, "sort_by", null)

    if (currSortKey === sortKey) {
      const currSortOrder = get(queryParams, "order", null)

      if (currSortOrder === SortOrders.ASC) {
        nextSort.order = SortOrders.DESC
      } else {
        nextSort.order = SortOrders.ASC
      }
    }

    setQueryParams(nextSort, false)
  }

  getCurrentSortKey(): string {
    return get(this.props.queryParams, "sort_by", "")
  }

  getCurrentSortOrder(): SortOrders {
    return get(this.props.queryParams, "order", "")
  }

  getCurrentPage(): number {
    return get(this.props.queryParams, "page", 1)
  }

  getChildProps(): IChildProps {
    const { requestState, data, error, inFlight, queryParams } = this.props

    return {
      queryParams,
      requestState,
      error,
      inFlight,
      data,
      updatePage: this.updatePage.bind(this),
      updateSort: this.updateSort.bind(this),
      updateSearch: this.updateSearch.bind(this),
      currentSortKey: this.getCurrentSortKey(),
      currentSortOrder: this.getCurrentSortOrder(),
      currentPage: this.getCurrentPage(),
    }
  }

  render() {
    return this.props.children(this.getChildProps())
  }
}

const ConnectedListRequest: React.SFC<IConnectedProps> = props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => (
      <Request requestFn={props.requestFn} shouldRequestOnMount={false}>
        {(requestChildProps: IRequestChildProps) => (
          <ListRequest
            {...props}
            data={requestChildProps.data}
            inFlight={requestChildProps.inFlight}
            requestState={requestChildProps.requestState}
            request={requestChildProps.request}
            error={requestChildProps.error}
            queryParams={queryParams}
            setQueryParams={setQueryParams}
          />
        )}
      </Request>
    )}
  </QueryParams>
)

export default ConnectedListRequest
