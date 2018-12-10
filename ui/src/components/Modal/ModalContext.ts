import { createContext } from "react"
import { IModalContext } from "../../.."

const ModalContext = createContext<IModalContext>({
  renderModal: () => {},
  unrenderModal: () => {},
})

export default ModalContext
