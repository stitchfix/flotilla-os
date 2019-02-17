import { createContext } from "react"
import { IFlotillaUIModalContext } from "../../types"

const ModalContext = createContext<IFlotillaUIModalContext>({
  renderModal: (modal: React.ReactNode) => {},
  unrenderModal: () => {},
})

export default ModalContext
