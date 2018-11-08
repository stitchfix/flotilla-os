import { createContext } from "react"

const PopupContext = createContext({
  renderPopup: () => {},
  unrenderPopup: () => {},
})

export default PopupContext
