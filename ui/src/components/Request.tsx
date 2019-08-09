import * as React from "react"
import { AxiosError } from "axios"

export enum RequestStatus {
  READY = "READY",
  NOT_READY = "NOT_READY",
  ERROR = "ERROR",
}

export type Props<ResponseType, ArgsType> = {
  children: (props: ChildProps<ResponseType, ArgsType>) => React.ReactNode
  requestFn: (args: ArgsType) => Promise<ResponseType>
  initialRequestArgs: ArgsType
  shouldRequestOnMount: boolean
  onSuccess?: (res: ResponseType) => void
  onFailure?: (error: any) => void
}

export type State<ResponseType> = {
  requestStatus: RequestStatus
  data: ResponseType | null
  isLoading: boolean
  error: AxiosError | null
}

export type ChildProps<ResponseType, ArgsType> = State<ResponseType> & {
  request: (opts: ArgsType) => void
}

class Request<ResponseType, ArgsType> extends React.Component<
  Props<ResponseType, ArgsType>,
  State<ResponseType>
> {
  static defaultProps = {
    shouldRequestOnMount: true,
    initialRequestArgs: null,
  }

  state = {
    requestStatus: RequestStatus.NOT_READY,
    data: null,
    isLoading: false,
    error: null,
  }

  componentDidMount() {
    if (this.props.shouldRequestOnMount) {
      this.request(this.props.initialRequestArgs)
    }
  }

  request(args: ArgsType): void {
    const { requestFn, onSuccess, onFailure } = this.props

    this.setState({ isLoading: true })

    requestFn(args)
      .then((data: ResponseType) => {
        this.setState({
          data,
          isLoading: false,
          requestStatus: RequestStatus.READY,
          error: null,
        })
        if (onSuccess) onSuccess(data)
      })
      .catch((error: AxiosError) => {
        this.setState({
          isLoading: false,
          requestStatus: RequestStatus.ERROR,
          error,
        })
        if (onFailure) onFailure(error)
      })
  }

  getChildProps = () => ({
    ...this.state,
    request: this.request.bind(this),
  })

  render() {
    return this.props.children(this.getChildProps())
  }
}

export default Request
