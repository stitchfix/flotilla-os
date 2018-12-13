import { createContext } from "react"
import { IFlotillaUIRunContext, flotillaUIRequestStates } from "../../.."

const RunContext = createContext<IFlotillaUIRunContext>({
  data: null,
  inFlight: false,
  error: false,
  requestState: flotillaUIRequestStates.NOT_READY,
  runID: "",
})

export default RunContext
