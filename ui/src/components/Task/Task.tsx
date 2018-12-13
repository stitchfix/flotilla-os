import * as React from "react"
import { isEqual } from "lodash"
import api from "../../api"
import TaskContext from "./TaskContext"
import { requestStates, IFlotillaTaskDefinition } from "../../.."

interface ITaskProps {
  definitionID: string
}

interface ITaskState {
  inFlight: boolean
  error: any
  data: IFlotillaTaskDefinition | null
  requestState: requestStates
}

class Task extends React.PureComponent<ITaskProps, ITaskState> {
  state = {
    inFlight: false,
    error: false,
    data: null,
    requestState: requestStates.NOT_READY,
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
    this.setState({ inFlight: false, error: false })

    api
      .getTask({ definitionID: this.props.definitionID })
      .then(data => {
        this.setState({
          inFlight: false,
          data,
          error: false,
          requestState: requestStates.READY,
        })
      })
      .catch(error => {
        this.setState({
          inFlight: false,
          error,
          requestState: requestStates.ERROR,
        })
      })
  }

  getCtx() {
    const { definitionID } = this.props
    return {
      ...this.state,
      definitionID,
      requestData: this.requestData.bind(this),
    }
  }

  render() {
    const { children } = this.props

    return (
      <TaskContext.Provider value={this.getCtx()}>
        {children}
      </TaskContext.Provider>
    )
  }
}

export default Task
