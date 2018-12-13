import * as React from "react"
import { withRouter, RouteComponentProps } from "react-router-dom"
import { isEmpty } from "lodash"
import qs from "qs"

interface IQueryParams {
  queryParams: any
  setQueryParams: (query: object, shouldReplace: boolean) => void
}

interface IUnwrappedQueryParamsProps extends IQueryParamsProps {
  push: (args: any) => void
  replace: (args: any) => void
  search: string
}

interface IQueryParamsProps {
  children: (opts: IQueryParams) => React.ReactNode
}

class UnwrappedQueryParams extends React.PureComponent<
  IUnwrappedQueryParamsProps
> {
  getQuery = (): any => {
    const { search } = this.props

    if (search.length > 0) {
      return qs.parse(search.slice(1))
    }

    return {}
  }

  setQuery = (query: object, shouldReplace = false): void => {
    const { replace, push } = this.props

    const next = qs.stringify(
      this.filterEmptyValues({
        ...this.getQuery(),
        ...query,
      }),
      { indices: false }
    )

    if (shouldReplace) {
      replace({ search: next })
    } else {
      push({ search: next })
    }
  }

  filterEmptyValues = (values: { [k: string]: any }): { [k: string]: any } =>
    Object.keys(values).reduce((acc: { [k: string]: any }, key: string): {
      [k: string]: any
    } => {
      if (!isEmpty(values[key])) {
        acc[key] = values[key]
      }

      return acc
    }, {})

  render() {
    return this.props.children({
      queryParams: this.getQuery(),
      setQueryParams: this.setQuery,
    })
  }
}

export default withRouter((props: RouteComponentProps & IQueryParamsProps) => (
  <UnwrappedQueryParams
    children={props.children}
    search={props.location.search}
    push={props.history.push}
    replace={props.history.replace}
  />
))
