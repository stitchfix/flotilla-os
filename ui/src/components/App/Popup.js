import React, { createContext, Component } from "react"
import PropTypes from "prop-types"

export const PopupContext = createContext({
  renderPopup: () => {},
  unrenderPopup: () => {},
})

class PopupContainer extends Component {
  state = {
    isVisible: false,
    popup: null,
  }

  renderPopup = popup => {
    this.setState({ isVisible: true, popup })
  }

  unrenderPopup = () => {
    this.setState({ isVisible: false, popup: null })
  }

  getCtx() {
    return {
      renderPopup: this.renderPopup,
      unrenderPopup: this.unrenderPopup,
    }
  }

  render() {
    const { popup, isVisible } = this.state

    return (
      <PopupContext.Provider value={this.getCtx()}>
        {!!isVisible && popup}
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
