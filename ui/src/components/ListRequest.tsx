import * as React from "react"
import { get, isEqual, isEmpty, Omit } from "lodash"
import Request, { ChildProps as RequestChildProps } from "./Request"
import QueryParams, { ChildProps as QueryChildProps } from "./QueryParams"
import { SortOrder } from "../types"

const DEFAULT_PROPS = {
  initialQuery: { page: 1 },
}

export type Props<Response, Args> = RequestChildProps<Response, Args> &
  QueryChildProps &
  Pick<
    ConnectedProps<Response, Args>,
    "children" | "initialQuery" | "getRequestArgs"
  >

export type ChildProps<Response, Args> = Omit<
  RequestChildProps<Response, Args>,
  "request"
> & {
  updateSort: (sortKey: string) => void
  updatePage: (n: number) => void
  updateFilter: (key: string, value: any) => void
  currentPage: number
  currentSortKey: string
  currentSortOrder: SortOrder
  query: any
}

export class ListRequest<Response, Args> extends React.Component<
  Props<Response, Args>
> {
  static defaultProps = DEFAULT_PROPS

  componentDidMount() {
    // Read query to see if relevant parameters are set
    if (isEmpty(this.props.query)) {
      this.props.setQuery(this.props.initialQuery, true)
    } else {
      this.request()
    }
  }

  componentDidUpdate(prevProps: Props<Response, Args>) {
    if (!isEqual(prevProps.query, this.props.query)) {
      this.request()
    }
  }

  request() {
    const { request, getRequestArgs, query } = this.props
    request(getRequestArgs(query))
  }

  /**
   * Updates the query's `sort_by` and `order` keys.
   * @param sortKey - the key to sort by
   */
  updateSort(sortKey: string): void {
    const { query, setQuery } = this.props
    const currSortKey = get(query, "sort_by", null)

    if (currSortKey === sortKey) {
      const currSortOrder = get(query, "order", null)

      if (currSortOrder === SortOrder.ASC) {
        setQuery({
          ...this.props.query,
          page: 1,
          sort_by: sortKey,
          order: SortOrder.DESC,
        })
      } else {
        setQuery({
          ...this.props.query,
          page: 1,
          sort_by: sortKey,
          order: SortOrder.ASC,
        })
      }
    } else {
      setQuery({
        ...this.props.query,
        page: 1,
        sort_by: sortKey,
        order: SortOrder.ASC,
      })
    }
  }

  /**
   * @param n - page number
   */
  updatePage(n: number): void {
    this.props.setQuery({ ...this.props.query, page: n })
  }

  /**
   * @param key - the filter's key, e.g. "alias"
   * @param value - the filter's value, e.g. "etl" or ["a", "b"]
   */
  updateFilter(key: string, value: any): void {
    this.props.setQuery({ ...this.props.query, page: 1, [key]: value })
  }

  getChildProps(): ChildProps<Response, Args> {
    return {
      requestStatus: this.props.requestStatus,
      data: this.props.data,
      isLoading: this.props.isLoading,
      error: this.props.error,
      updateSort: this.updateSort.bind(this),
      updatePage: this.updatePage.bind(this),
      updateFilter: this.updateFilter.bind(this),
      currentPage: Number(get(this.props.query, "page", 1)),
      currentSortKey: get(this.props.query, "sort_by", ""),
      currentSortOrder: get(this.props.query, "order", ""),
      query: this.props.query,
    }
  }

  render() {
    return this.props.children(this.getChildProps())
  }
}

type ConnectedProps<Response, Args> = {
  children: (props: ChildProps<Response, Args>) => React.ReactNode
  requestFn: (args: Args) => Promise<Response>
  initialQuery: object
  getRequestArgs: (query: object) => Args
}

class ConnectedListRequest<Response, Args> extends React.Component<
  ConnectedProps<Response, Args>
> {
  static defaultProps = DEFAULT_PROPS
  render() {
    const { requestFn, initialQuery, getRequestArgs, children } = this.props
    return (
      <Request requestFn={requestFn} shouldRequestOnMount={false}>
        {requestProps => (
          <QueryParams>
            {({ query, setQuery }) => (
              <ListRequest
                {...requestProps}
                query={query}
                setQuery={setQuery}
                initialQuery={initialQuery}
                getRequestArgs={getRequestArgs}
              >
                {children}
              </ListRequest>
            )}
          </QueryParams>
        )}
      </Request>
    )
  }
}

export default ConnectedListRequest
