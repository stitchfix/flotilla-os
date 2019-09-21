import * as React from "react"
import * as qs from "qs"
import { withRouter, RouteComponentProps } from "react-router-dom"

type Props = RouteComponentProps & {
  children: (props: ChildProps) => React.ReactNode
}

export type ChildProps = {
  query: object
  setQuery: (query: object, shouldReplace?: boolean) => void
}

export class QueryParams extends React.Component<Props> {
  setQuery(query: object, shouldReplace?: boolean): void {
    const { history } = this.props

    if (shouldReplace === true) {
      history.replace({ search: qs.stringify(query, { indices: false }) })
    } else {
      history.push({ search: qs.stringify(query, { indices: false }) })
    }
  }

  getQuery(): object {
    const { location } = this.props

    if (location.search.length > 0) {
      return qs.parse(location.search.substr(1))
    }

    return {}
  }

  getChildProps(): ChildProps {
    return {
      query: this.getQuery(),
      setQuery: this.setQuery.bind(this),
    }
  }

  render() {
    return this.props.children(this.getChildProps())
  }
}

export default withRouter(QueryParams)
