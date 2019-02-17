import { createContext } from "react"
import { flotillaUIRequestStates, IFlotillaUITaskContext } from "../../types"

const TaskContext = createContext<IFlotillaUITaskContext>({
  data: null,
  inFlight: false,
  error: false,
  requestState: flotillaUIRequestStates.NOT_READY,
  definitionID: "",
  requestData: () => {},
})

export default TaskContext
