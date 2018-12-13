import { createContext } from "react"
import { IFlotillaUIPopupContext } from "../../.."

const PopupContext = createContext<IFlotillaUIPopupContext>({
  renderPopup: () => {},
  unrenderPopup: () => {},
})

export default PopupContext
