import { createContext } from "react"
import { IFlotillaUIPopupContext } from "../../types"

const PopupContext = createContext<IFlotillaUIPopupContext>({
  renderPopup: () => {},
  unrenderPopup: () => {},
})

export default PopupContext
