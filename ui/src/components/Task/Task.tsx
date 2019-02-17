import * as React from "react"
import { get, isEqual, isEmpty } from "lodash"
import api from "../../api"
import TaskContext from "./TaskContext"
import {
  flotillaUIRequestStates,
  IFlotillaTaskDefinition,
  IFlotillaAPIError,
  IFlotillaUITaskContext,
} from "../../types"
import Loader from "../styled/Loader"

interface ITaskProps {
  definitionID: string
  alias?: string
  shouldRequestByAlias: boolean
}

interface ITaskState {
  inFlight: boolean
  error: any
  data: IFlotillaTaskDefinition | null
  requestState: flotillaUIRequestStates
}

class Task extends React.PureComponent<ITaskProps, ITaskState> {
  state = {
    inFlight: false,
    error: false,
    data: null,
    requestState: flotillaUIRequestStates.NOT_READY,
  }

  componentDidMount() {
    this.requestData()
  }

  componentDidUpdate(prevProps: ITaskProps) {
    if (!isEqual(prevProps.definitionID, this.props.definitionID)) {
      this.requestData()
    }
  }

  requestData(): void {
    const { definitionID, alias, shouldRequestByAlias } = this.props
    this.setState({ inFlight: false, error: false })

    if (!!shouldRequestByAlias && !!alias) {
      api
        .getTaskByAlias({ alias })
        .then(this.handleResponse)
        .catch(this.handleResponse)

      return
    }

    api
      .getTask({ definitionID })
      .then(this.handleResponse)
      .catch(this.handleResponse)
  }

  handleResponse = (data: IFlotillaTaskDefinition): void => {
    this.setState({
      inFlight: false,
      data,
      error: false,
      requestState: flotillaUIRequestStates.READY,
    })
  }

  handleError = (error: IFlotillaAPIError): void => {
    this.setState({
      inFlight: false,
      error,
      requestState: flotillaUIRequestStates.ERROR,
    })
  }

  getCtx = (): IFlotillaUITaskContext => {
    const { definitionID, shouldRequestByAlias } = this.props

    let ret = {
      ...this.state,
      definitionID,
      requestData: this.requestData.bind(this),
    }

    if (shouldRequestByAlias) {
      ret.definitionID = get(this.state, ["data", "definition_id"], null)
    }

    return ret
  }

  render() {
    const { children, shouldRequestByAlias } = this.props

    if (shouldRequestByAlias && isEmpty(this.state.data)) {
      return <Loader />
    }

    return (
      <TaskContext.Provider value={this.getCtx()}>
        {children}
      </TaskContext.Provider>
    )
  }
}

export default Task
