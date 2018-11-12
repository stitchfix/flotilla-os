import React, { Component } from "react"
import PropTypes from "prop-types"
import styled from "styled-components"
import PopupContext from "./PopupContext"
import { Z_INDICES } from "../../constants/styles"
import colors from "../../constants/colors"
import intentTypes from "../../constants/intentTypes"

const POPUP_WINDOW_DISTANCE_PX = 48
const POPUP_WIDTH_PX = 320

const PopupPositioner = styled.div`
  position: fixed;
  bottom: ${POPUP_WINDOW_DISTANCE_PX}px;
  right: ${POPUP_WINDOW_DISTANCE_PX}px;
  z-index: ${Z_INDICES.POPUP};
  background: ${colors.black[2]};
  width: ${POPUP_WIDTH_PX}px;
`

const PopupContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: flex-start;
  align-items: stretch;
  width: 100%;
  height: 100%;
`

const PopupTitle = styled.div``
const PopupBody = styled.div``

class Popup extends Component {
  componentDidMount() {
    const { shouldAutohide, unrenderPopup, visibleDuration } = this.props

    if (!!shouldAutohide) {
      window.setTimeout(() => {
        unrenderPopup()
      }, visibleDuration)
    }
  }

  render() {
    const { body, title } = this.props
    return (
      <PopupPositioner>
        <PopupContainer>
          {!!title && <PopupTitle>{title}</PopupTitle>}
          {!!body && <PopupBody>{body}</PopupBody>}
        </PopupContainer>
      </PopupPositioner>
    )
  }
}

Popup.propTypes = {
  body: PropTypes.node,
  intent: PropTypes.oneOf(Object.values(intentTypes)),
  shouldAutohide: PropTypes.bool.isRequired,
  title: PropTypes.node,
  unrenderPopup: PropTypes.func.isRequired,
  visibleDuration: PropTypes.number.isRequired,
}

Popup.defaultProps = {
  shouldAutohide: true,
  visibleDuration: 5000,
}

export default props => (
  <PopupContext.Consumer>
    {pCtx => <Popup {...props} unrenderPopup={pCtx.unrenderPopup} />}
  </PopupContext.Consumer>
)
