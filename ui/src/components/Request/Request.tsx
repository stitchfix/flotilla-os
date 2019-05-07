import * as React from "react"
import { isArray } from "lodash"
import { flotillaUIRequestStates } from "../../types"

type RequestArgs = { [key: string]: any }
type RequestFn = (args?: any) => Promise<any>

export interface IProps {
  shouldRequestOnMount: boolean
  initialRequestArgs?: RequestArgs | RequestArgs[]
  requestFn: RequestFn | RequestFn[]
  children: (props: IChildProps) => React.ReactNode
}

export interface IState {
  inFlight: boolean
  data: any
  requestState: flotillaUIRequestStates
  error: any
}

export interface IChildProps extends IState {
  request: (args?: any) => void
}

/**
 * This component takes a requestFn prop (any function that returns a promise)
 * then calls its `children` prop with the promise's resolved value, along with
 * other useful attributes.
 */
class Request extends React.PureComponent<IProps, IState> {
  static defaultProps: Partial<IProps> = {
    shouldRequestOnMount: true,
  }

  state = {
    inFlight: false,
    data: null,
    requestState: flotillaUIRequestStates.NOT_READY,
    error: false,
  }

  componentDidMount() {
    const { shouldRequestOnMount, initialRequestArgs } = this.props

    if (shouldRequestOnMount) {
      this.request(initialRequestArgs)
    }
  }

  request(requestArgs?: RequestArgs | RequestArgs[]): void {
    const { requestFn } = this.props
    this.setState({ inFlight: true })

    if (isArray(requestFn) && isArray(requestArgs)) {
      Promise.all(requestFn.map((fn, i) => fn(requestArgs[i])))
        .then(data => {
          this.setState({
            data,
            inFlight: false,
            error: false,
            requestState: flotillaUIRequestStates.READY,
          })
        })
        .catch(error => {
          this.setState({
            inFlight: false,
            error,
            requestState: flotillaUIRequestStates.ERROR,
          })
        })
    } else if (!isArray(requestFn) && !isArray(requestArgs)) {
      requestFn(requestArgs)
        .then(data => {
          this.setState({
            data,
            inFlight: false,
            error: false,
            requestState: flotillaUIRequestStates.READY,
          })
        })
        .catch(error => {
          this.setState({
            inFlight: false,
            error,
            requestState: flotillaUIRequestStates.ERROR,
          })
        })
    } else {
      const errMsg =
        "The `requestFn` and `requestArgs` props passed to <Request> must either both be arrays or non-arrays."
      this.setState({
        inFlight: false,
        error: errMsg,
        requestState: flotillaUIRequestStates.ERROR,
      })
      console.error(errMsg)
    }
  }

  render() {
    return this.props.children({
      ...this.state,
      request: this.request.bind(this),
    })
  }
}

export default Request
