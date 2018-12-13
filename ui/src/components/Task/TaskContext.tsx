import { createContext } from "react"
import { requestStates, IFlotillaUITaskContext } from "../../.."

const TaskContext = createContext<IFlotillaUITaskContext>({
  data: null,
  inFlight: false,
  error: false,
  requestState: requestStates.NOT_READY,
  definitionID: "",
  requestData: () => {},
})

export default TaskContext
