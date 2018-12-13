import * as React from "react"
import PopupContext from "./PopupContext"
import Popup from "./Popup"
import { IFlotillaUIPopupProps } from "../../.."

interface IPopupContainerState {
  isVisible: boolean
  popupProps: IFlotillaUIPopupProps | undefined
}

class PopupContainer extends React.Component<{}, IPopupContainerState> {
  state = {
    isVisible: false,
    popupProps: undefined,
  }

  renderPopup = (popupProps: IFlotillaUIPopupProps) => {
    this.setState({ isVisible: true, popupProps })
  }

  unrenderPopup = () => {
    this.setState({ isVisible: false, popupProps: undefined })
  }

  getCtx() {
    return {
      renderPopup: this.renderPopup,
      unrenderPopup: this.unrenderPopup,
    }
  }

  render() {
    const { popupProps, isVisible } = this.state

    return (
      <PopupContext.Provider value={this.getCtx()}>
        {!!isVisible && popupProps !== undefined && <Popup {...popupProps} />}
        {this.props.children}
      </PopupContext.Provider>
    )
  }
}

export default PopupContainer
