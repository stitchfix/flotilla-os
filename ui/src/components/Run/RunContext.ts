import { createContext } from "react"
import { IFlotillaRunContext, requestStates } from "../../.."

const RunContext = createContext<IFlotillaRunContext>({
  data: null,
  inFlight: false,
  error: false,
  requestState: requestStates.NOT_READY,
  runID: "",
})

export default RunContext
