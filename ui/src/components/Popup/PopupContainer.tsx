import React, { Component } from "react"
import PropTypes from "prop-types"
import PopupContext from "./PopupContext"
import Popup from "./Popup"
import { IPopupProps } from "../../.."

interface IPopupContainerState {
  isVisible: boolean
  popupProps: IPopupProps | undefined
}

class PopupContainer extends Component<{}, IPopupContainerState> {
  state = {
    isVisible: false,
    popupProps: undefined,
  }

  renderPopup = (popupProps: IPopupProps) => {
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
