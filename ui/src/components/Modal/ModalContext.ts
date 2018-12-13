import { createContext } from "react"
import { IModalContext } from "../../.."

const ModalContext = createContext<IModalContext>({
  renderModal: (modal: React.ReactNode) => {},
  unrenderModal: () => {},
})

export default ModalContext
