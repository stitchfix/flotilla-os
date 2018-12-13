import { createContext } from "react"
import { IFlotillaUIModalContext } from "../../.."

const ModalContext = createContext<IFlotillaUIModalContext>({
  renderModal: (modal: React.ReactNode) => {},
  unrenderModal: () => {},
})

export default ModalContext
