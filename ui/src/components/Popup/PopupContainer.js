import React, { createContext, Component } from "react"
import PropTypes from "prop-types"
import PopupContext from "./PopupContext"
import Popup from "./Popup"

class PopupContainer extends Component {
  state = {
    isVisible: false,
    popupProps: null,
  }

  renderPopup = popupProps => {
    this.setState({ isVisible: true, popupProps })
  }

  unrenderPopup = () => {
    this.setState({ isVisible: false, popupProps: null })
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
        {!!isVisible && <Popup {...popupProps} />}
        {this.props.children}
      </PopupContext.Provider>
    )
  }
}

PopupContainer.displayName = "PopupContainer"

PopupContainer.propTypes = {
  children: PropTypes.node,
}

export default PopupContainer
