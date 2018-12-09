import { createContext } from "react"
import { IPopupContext } from "../../.."

const PopupContext = createContext<IPopupContext>({
  renderPopup: () => {},
  unrenderPopup: () => {},
})

export default PopupContext
