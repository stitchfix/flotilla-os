import { createContext } from "react"
import * as requestStateTypes from "../../helpers/requestStateTypes"

const TaskContext = createContext({
  data: {},
  inFlight: false,
  error: false,
  requestState: requestStateTypes.NOT_READY,
  definitionID: null,
  requestData: () => {},
})

export default TaskContext
