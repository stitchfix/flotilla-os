import { createContext } from "react"
import * as requestStateTypes from "../../helpers/requestStateTypes"

const RunContext = createContext({
  data: {},
  inFlight: false,
  error: false,
  requestState: requestStateTypes.NOT_READY,
  definitionID: null,
  runID: null,
})

export default RunContext
