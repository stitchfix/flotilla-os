import { createContext } from "react"

const ModalContext = createContext({
  renderModal: () => {},
  unrenderModal: () => {},
})

export default ModalContext
